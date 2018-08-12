package git

import (
	"testing"

	"github.com/valer-cara/mgo/pkg/testutils"
)

func TestInexistentGitBackend(t *testing.T) {
	_, err := NewGit(909090, "whatever")
	if err == nil {
		t.Fatal("Should not allow initialization with inexsitent git backend")
	}
}

func TestWithExternalBackendGoodRepo(t *testing.T) {
	repo, _ := testutils.CreateTestRepoWithOrigin(t)
	_, err := NewGit(BACKEND_EXTERNAL, repo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWithExternalBackendBadRepo(t *testing.T) {
	_, err := NewGit(BACKEND_EXTERNAL, "/tmp/no-repo-here-nope")
	if err == nil {
		t.Fatal("non-git-repo should have returned an error")
	}
}
