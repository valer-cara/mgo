package deploy

import (
	"testing"

	"github.com/valer-cara/mgo/pkg/git"
	"github.com/valer-cara/mgo/pkg/manifest"
)

func TestNewDeploy(t *testing.T) {
	gitService, _ := git.NewGit(git.BACKEND_FAKE, "whatevs")
	x := NewDeploy(gitService, &FakeUpdater{}, &DeployOptions{
		Author:      "Ronaldo",
		TriggerRepo: "git.kernel.org",
		Image: manifest.HeaderImage{
			Repository: "quay.io/foobar",
			Tag:        "beta",
		},
		Cluster: "fake-cluster-here",
	})

	err := x.Create()
	if err != nil {
		t.Fatal("Expected no errors when creating with a fake git backend:", err)
	}
}

func TestDeployUpdatesCorrespondingImages(t *testing.T) {

}
