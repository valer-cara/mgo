package deploy

import (
	"testing"

	"github.com/valer-cara/mgo/pkg/testutils"
)

func TestOnlyChangeChartSection(t *testing.T) {
	repo := testutils.CreateTestRepoFromSample(t, "../../tests/minimal-gitops-repo")
	t.Log("Testing repo", repo)

	updater := MyUpdater{}
	err := updater.Update(repo, &DeployOptions{
		Author:      "Ronaldo",
		TriggerRepo: "github.com/a/repo1",
		Image: DeployOptionsImage{
			Repository: "quay.io/foobar",
			Tag:        "beta",
		},
		Cluster: "myprodcluster",
	})

	if err != nil {
		t.Fatal(err)
	}
}
