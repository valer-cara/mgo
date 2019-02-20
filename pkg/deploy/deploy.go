package deploy

import (
	"fmt"

	"github.com/valer-cara/mgo/pkg/git"
	"github.com/valer-cara/mgo/pkg/util"
)

type Deploy struct {
	options        *DeployOptions
	gitService     *git.Git
	updaterService Updater
}

type DeployOptionsImage struct {
	Repository string `json:"repository"`
	Tag        string `json:"tag"`
}

type DeployOptions struct {
	// The URL of the repo that triggered this deploy
	TriggerRepo string `json:"triggerRepo"`

	// Who's to blame for all this
	Author string `json:"author"`

	// The docker image built by the CI pipeline for this deploy
	Image DeployOptionsImage

	// The target cluster for this deploy
	Cluster string `json:"cluster"`
}

func (d *DeployOptions) String() string {
	return fmt.Sprintf("triggerRepo: %s, author: %s, cluster: %s, image: %s:%s",
		d.TriggerRepo,
		d.Author,
		d.Cluster,
		d.Image.Repository,
		d.Image.Tag,
	)
}

func NewDeploy(gitService *git.Git, updaterService Updater, options *DeployOptions) *Deploy {
	return &Deploy{
		options:        options,
		gitService:     gitService,
		updaterService: updaterService,
	}
}

func (d *Deploy) Create() error {
	return d.doCreate()
}

func (d *Deploy) msg() string {
	return fmt.Sprintf("Deploy: %s:%s to %s by %s", d.options.Image.Repository, d.options.Image.Tag, d.options.Cluster, d.options.Author)
}

func (d *Deploy) doCreate() error {
	err := util.CallFunctions(
		func() error {
			return d.updaterService.Update(d.gitService.Root(), d.options)
		},
		func() error { return d.gitService.AddAll() },
		func() error { return d.gitService.Commit(d.msg()) },
	)

	return err
}
