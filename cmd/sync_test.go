package cmd

import (
	"testing"

	"github.com/valer-cara/mgo/pkg/testutils"
)

// TODO: think about user vs. fresh helm home dir
// Test sync with user's helm dir
func TestSyncCmdUserHelm(t *testing.T) {
}

// Test sync with a fresh helm home dir
func TestSyncCmdFreshHelm(t *testing.T) {
	repo := testutils.CreateTestRepoFromSample(t, "../tests/minikube-gitops-repo")
	t.Log("Test repo at: ", repo)

	err := testutils.PrepareArgs(t, syncCmd, []string{
		"--gitops-repo=" + repo,
		"--cluster=minikube",
	})
	if err != nil {
		t.Fatal("Error parsing arguments:", err)
	}

	// Leftovers from rootcmd
	initConfig()

	err = doSync()
	if err != nil {
		t.Fatal("Expected successful sync: ", err)
	}
}
