package cmd

import (
	"testing"

	"github.com/valer-cara/mgo/pkg/testutils"
)

func TestDeploy(t *testing.T) {
	repo := testutils.CreateTestRepoFromSample(t, "../tests/minimal-gitops-repo")

	t.Log("Test repo at: ", repo)

	err := testutils.PrepareArgs(t, deployCmd, []string{
		"--gitops-repo=" + repo,
		"--cluster=myprodcluster",
		"--author=Freddie",
		"--source=github.com/a/repo1",
	})
	if err != nil {
		t.Fatal("Error parsing arguments:", err)
	}

	err = doDeploy()
	if err != nil {
		t.Fatal("Expected successful deploy: ", err)
	}
}

func TestDeployBadRepo(t *testing.T) {
	err := testutils.PrepareArgs(t, deployCmd, []string{
		"--gitops-repo=/tmp/THIS_IS_NO_REPO_FRIENDS",
		"--cluster=myprodcluster",
		"--author=Freddie",
		"--source=github.com/a/repo1",
	})
	if err != nil {
		t.Fatal("Error parsing arguments:", err)
	}

	err = doDeploy()
	if err == nil {
		t.Fatal("Expected failed command on nonexisting repo.")
	}
}

func TestUpdatesCommited(t *testing.T) {
	t.Skip("To be implemented")
}
