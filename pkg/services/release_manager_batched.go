package services

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/valer-cara/mgo/pkg/async"
	btch "github.com/valer-cara/mgo/pkg/batcher"
	"github.com/valer-cara/mgo/pkg/config"
	"github.com/valer-cara/mgo/pkg/deploy"
	"github.com/valer-cara/mgo/pkg/git"
	"github.com/valer-cara/mgo/pkg/helm"
	clusterSync "github.com/valer-cara/mgo/pkg/sync"
	"github.com/valer-cara/mgo/pkg/util"

	"github.com/avast/retry-go"
)

// This ReleaseManager creates deployment commits in the gitops repo
// for each incoming deploy request.
//
// Based on a batching processor that:
// 1. updates the gitops repo to the latest as a PreHook
// 2. goes over all queued deployments and creates a commit for each
// 3. syncs all affected clusters and returns the corresponding statuses
//
type ReleaseManagerBatched struct {
	options *ReleaseManagerBatchedOptions

	// dependent services
	gitService   *git.Git
	syncServices map[string]*clusterSync.Sync
	helmServices map[string]helm.HelmService

	// batcher
	batcher       *btch.Batcher
	chanBatchDone chan bool
	chanBatchErr  chan error

	clusterSyncWaitlists map[string]*async.Waitlist
}

type ReleaseManagerBatchedOptions struct {
	GitopsRepo, KubeConfig, HelmHome string
	DryRun                           bool
}

type kubeconfigClusters struct {
	Clusters []struct {
		Name string
	}
}

func NewReleaseManagerBatched(opts *ReleaseManagerBatchedOptions) *ReleaseManagerBatched {
	chanBatchDone, chanBatchErr := make(chan bool), make(chan error)

	relMgr := &ReleaseManagerBatched{
		options:       opts,
		syncServices:  make(map[string]*clusterSync.Sync),
		helmServices:  make(map[string]helm.HelmService),
		chanBatchDone: chanBatchDone,
		chanBatchErr:  chanBatchErr,

		// maps clusterName -> array of async results
		clusterSyncWaitlists: make(map[string]*async.Waitlist),
	}

	batcherOpts := &btch.BatcherOptions{
		// PreBatch hook summary: sync local gitops repo
		// PostBatch hook summary: sync updated state in gitops repo to cluster
		PreBatch:  relMgr.createPreBatchHook(),
		PostBatch: relMgr.createPostBatchHook(),
		Done:      chanBatchDone,
		Err:       chanBatchErr,
	}

	relMgr.batcher = btch.NewBatcher(batcherOpts)

	return relMgr
}

func (r *ReleaseManagerBatched) Init() error {
	// Init git service
	gitService, err := git.NewGit(git.BACKEND_EXTERNAL, r.options.GitopsRepo)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot initialize git service on %s", r.options.GitopsRepo))
	}
	r.gitService = gitService

	if err := r.initPerClusterServices(); err != nil {
		return errors.New(fmt.Sprintf("Cannot determine available kubernetes clusters: %v", err))
	}

	go r.batcher.Start()
	go r.monitorBatch()

	return nil
}

func (r *ReleaseManagerBatched) initPerClusterServices() error {
	file, err := ioutil.ReadFile(r.options.KubeConfig)
	if err != nil {
		return err
	}

	kc := kubeconfigClusters{}
	err = yaml.Unmarshal(file, &kc)
	if err != nil {
		return errors.New(fmt.Sprintf("YAML unmarshal %s, %v", r.options.KubeConfig, err))
	}

	// For each context (cluster) defined in kubeconfig, initialize a helm/sync
	// service pair set for that specific context
	for _, cluster := range kc.Clusters {
		helmService, err := r.initHelmService(cluster.Name)
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot initialize helm service for cluster %s: %v", cluster.Name, err))
		}

		r.helmServices[cluster.Name] = helmService
		r.syncServices[cluster.Name] = clusterSync.NewSync(r.options.GitopsRepo, cluster.Name, helmService)
		r.clusterSyncWaitlists[cluster.Name] = async.NewWaitlist()
	}

	return nil
}

func (r *ReleaseManagerBatched) initHelmService(cluster string) (helm.HelmService, error) {
	helmService := helm.NewHelmCmd(&helm.HelmCmdOptions{
		DryRun:       r.options.DryRun,
		HelmHome:     r.options.HelmHome,
		Kubeconfig:   r.options.KubeConfig,
		Kubecontext:  cluster,
		Repositories: config.Global.Helm.Repositories,
	})
	if err := helmService.Init(); err != nil {
		return nil, err
	}

	return helmService, nil
}

func (r *ReleaseManagerBatched) RequestRelease(dopts *deploy.DeployOptions) error {
	if r.syncServices[dopts.Cluster] == nil {
		return errors.New("Requested cluster is not managed by this instance of mygitops. Check `cluster` parameter.")
	}

	job := r.newDeployJob(dopts)
	chanJobDone, chanJobError := make(chan bool), make(chan error)
	clusterSyncResult := async.NewResult()

	r.batcher.Queue(job, chanJobDone, chanJobError)

	// await deployment committed
	select {
	case <-chanJobDone:
		r.clusterSyncWaitlists[dopts.Cluster].Add(clusterSyncResult)
	case err := <-chanJobError:
		return err
	}

	// await deployment synced to cluster
	select {
	case <-clusterSyncResult.Done:
		return nil
	case err := <-clusterSyncResult.Err:
		return err
	}

	return nil
}

// Update local gitops repository, preparing for new deploy-related edits
func (r *ReleaseManagerBatched) createPreBatchHook() func() error {
	return func() error {
		err := util.CallFunctions(
			func() error { return r.gitService.Fetch() },
			func() error { return r.gitService.Reset() }, // --hard origin master (or branch)
		)

		return err
	}
}

// Update local gitops repository, preparing for new deploy-related edits
func (r *ReleaseManagerBatched) createPostBatchHook() func() error {
	return func() error {
		// Push updates, and retry a few times while doing a `git pull -r`
		err := retry.Do(
			r.gitService.Push,
			retry.Attempts(4),
			retry.OnRetry(func(n uint, err error) {
				log.Warnln("Retrying git push: ", err)
				if err := r.gitService.Pull("-r"); err != nil {
					log.Warnln("Git pull -r failed: ", err)
				}
			}),
		)
		if err != nil {
			return err
		}

		for cluster, waitlist := range r.clusterSyncWaitlists {
			if !waitlist.IsEmpty() {
				log.Printf("Syncing cluster %s", cluster)

				err := r.syncCluster(cluster)
				if err != nil {
					waitlist.AllError(err)
					log.Errorf("Error syncing cluster %s: %s", cluster, err)
				} else {
					waitlist.AllDone()
					log.Printf("Done syncing cluster %s", cluster)
				}

				r.clusterSyncWaitlists[cluster].Clear()
			}
		}

		return nil
	}
}

func (r *ReleaseManagerBatched) syncCluster(cluster string) error {
	err := r.helmServices[cluster].UpdateRepos()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to update helm repos while syncing cluster %s: %s", cluster, err))
	}

	return r.syncServices[cluster].Sync()
}

func (r *ReleaseManagerBatched) monitorBatch() {
	for {
		select {
		case <-r.chanBatchDone:
			log.Println("Batch done")
		case err := <-r.chanBatchErr:
			log.Errorf("Batch error: %v", err)
		}
	}
}

func (r *ReleaseManagerBatched) newDeployJob(dopts *deploy.DeployOptions) btch.Job {
	return func() error {
		log.Debugln("NewDeploy:", dopts.String())

		dpl := deploy.NewDeploy(r.gitService, &deploy.MyUpdater{}, dopts)

		err := dpl.Create()
		if err != nil {
			return errors.New(fmt.Sprintf("Cannot create deployment: %v", err))
		}

		return nil
	}
}
