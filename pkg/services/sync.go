package services

import (
	"errors"
	"fmt"

	"github.com/valer-cara/mgo/pkg/config"
	"github.com/valer-cara/mgo/pkg/helm"
	"github.com/valer-cara/mgo/pkg/sync"
)

// TODO: too much repetition of gitopsRepo, helmHome, kubeconfig in this program.. please distill
type SyncService struct {
	gitopsRepo  string
	helmHome    string
	kubeconfig  string
	kubecontext string
	dryRun      bool

	helmService helm.HelmService
}

func NewSyncService(gitopsRepo, helmHome, kubeconfig, kubecontext string, dryRun bool) *SyncService {
	return &SyncService{
		gitopsRepo:  gitopsRepo,
		kubeconfig:  kubeconfig,
		kubecontext: kubecontext,
		helmHome:    helmHome,
		dryRun:      dryRun,
	}
}

func (ss *SyncService) Init() error {
	ss.helmService = helm.NewHelmCmd(&helm.HelmCmdOptions{
		DryRun:       ss.dryRun,
		HelmHome:     ss.helmHome,
		Kubeconfig:   ss.kubeconfig,
		Kubecontext:  ss.kubecontext,
		Repositories: config.Global.Helm.Repositories,
	})

	if err := ss.helmService.Init(); err != nil {
		return errors.New(fmt.Sprintf(
			"Cannot sync cluster: Cannot initialize helm service: %v",
			err,
		))
	}

	return nil
}

func (ss *SyncService) Execute() error {
	syncService := sync.NewSync(ss.gitopsRepo, ss.kubecontext, ss.helmService)

	err := syncService.Sync()
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Cannot sync cluster: %v",
			err,
		))
	}

	return nil
}

func (ss *SyncService) String() string {
	return fmt.Sprintf(
		"SyncService(cluster: %s, gitops repo: %s, kubeconfig: %s)",
		ss.kubecontext,
		ss.gitopsRepo,
		ss.kubeconfig,
	)
}
