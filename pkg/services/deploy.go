package services

import (
	"errors"
	"fmt"

	"github.com/valer-cara/mgo/pkg/deploy"
	"github.com/valer-cara/mgo/pkg/git"
)

// Wraps functionality to create a new deployment
type DeployService struct {
	dopts      *deploy.DeployOptions
	gitopsRepo string
}

func NewDeployService(gitopsRepo string, dopts *deploy.DeployOptions) *DeployService {
	return &DeployService{
		dopts:      dopts,
		gitopsRepo: gitopsRepo,
	}
}

func (ds *DeployService) Execute() error {
	gitService, err := git.NewGit(git.BACKEND_EXTERNAL, ds.gitopsRepo)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot initialize git service on %s", ds.gitopsRepo))
	}

	dpl := deploy.NewDeploy(gitService, &deploy.MyUpdater{}, ds.dopts)

	err = dpl.Create()
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot create deployment: %v", err))
	}

	return nil
}

func (ds *DeployService) String() string {
	return fmt.Sprintf("DeployService(cluster: %s, author: %s, image: %s:%s)",
		ds.dopts.Cluster,
		ds.dopts.Author,
		ds.dopts.Image.Repository,
		ds.dopts.Image.Tag,
	)
}
