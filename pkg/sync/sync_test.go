package sync

import (
	"github.com/valer-cara/mgo/pkg/helm"
	"github.com/valer-cara/mgo/pkg/testutils"
	"testing"
)

func TestSync(t *testing.T) {
	repo := testutils.CreateTestRepoFromSample(t, "../../tests/minimal-gitops-repo")

	helmService := helm.HelmFake{}

	x := NewSync(repo, "myprodcluster", &helmService)
	err := x.Sync()

	if err != nil {
		t.Fatalf("Cannot sync repo %s: %v", repo, err)
	}
}
