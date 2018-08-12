package sync

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/valer-cara/mgo/pkg/helm"
	"github.com/valer-cara/mgo/pkg/jobs"
	"github.com/valer-cara/mgo/pkg/manifest"
	"github.com/valer-cara/mgo/pkg/util"
)

type Sync struct {
	gitopsRepoRoot string
	cluster        string
	files          struct {
		raw    []string
		values []string
	}

	helmService helm.HelmService
}

func NewSync(gitopsRepoRoot string, cluster string, helmService helm.HelmService) *Sync {
	return &Sync{
		gitopsRepoRoot: gitopsRepoRoot,
		cluster:        cluster,
		helmService:    helmService,
	}
}

func (s *Sync) Sync() error {
	manifests, err := manifest.FindManifests(s.gitopsRepoRoot, s.cluster)
	if err != nil {
		return err
	}

	s.files.raw = manifests.Raw
	s.files.values = manifests.Helm

	if err := s.syncHelmManifsets(); err != nil {
		return err
	}

	if err := s.syncRawManifsets(); err != nil {
		return err
	}

	return nil
}

type syncJob struct {
	Release    *helm.HelmRelease
	ValueFiles []string
}

func (s *Sync) syncHelmManifsets() error {
	var (
		errs     []error
		syncJobs []interface{}
	)

	for _, path := range s.files.values {
		//log.Printf("Syncing %s...\n", path)

		header, err := manifest.ParseHeader(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		if err := header.Validate(); err != nil {
			errs = append(errs, errors.New("Manifest "+path+":"+err.Error()))
			continue
		}

		// XXX: quick hack to include plaintext secrets here.... need to use sops/helm-secrets instead...
		files := []string{path}

		secretsFile := strings.Replace(path, "-values.yaml", "-secrets.yaml", -1)
		if _, err := os.Stat(secretsFile); err == nil {
			files = append(files, secretsFile)
		}

		var job interface{} = syncJob{
			Release:    &header.HelmRelease,
			ValueFiles: files,
		}

		syncJobs = append(syncJobs, job)
	}

	if len(errs) > 0 {
		return util.AggregateErrors(errs)
	}

	errs = jobs.Parallel(func(job interface{}) error {
		j := job.(syncJob)

		err := s.helmService.SyncRelease(j.Release, j.ValueFiles)
		if err != nil {
			return err
		}

		return nil
	}, syncJobs, &jobs.ParallelOpts{MaxParallel: 15})

	if len(errs) > 0 {
		return util.AggregateErrors(errs)
	}

	return nil
}

func (s *Sync) syncRawManifsets() error {
	for _, path := range s.files.raw {
		log.Printf("TODO: Not yet syncing %s...", path)
	}
	return nil
}
