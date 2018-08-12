package testutils

import (
	"io/ioutil"
	"os/exec"
	"path"
	"path/filepath"
	"testing"

	"github.com/valer-cara/mgo/pkg/util"
)

// XXX: THESE ARE ACTUAL COMMANDS RUNNING ON YOUR SYSTEM. WATCH OUT!
// need to think of a safer way to do it?

// TODO: keep track of all test repos and remove them after testing

// Create a new repo and setup local user.name/user.email useful for committing to that repo
// Returns the path to the new repo
func CreateTestRepo(t *testing.T, extraArgs ...string) string {
	repo, err := ioutil.TempDir("/tmp/", "_mygitops-test-")
	if err != nil {
		t.Fatal(err)
	}

	envRepo := []string{
		"GIT_DIR=" + path.Join(repo, ".git"),
	}

	args := append([]string{"init"}, extraArgs...)

	err = util.RunCommands(envRepo,
		// Git init args...
		exec.Command("git", args...),

		// Setup author name/email on repo for further commits
		exec.Command("git", "config", "--local", "user.name", "MyGitops Robot"),
		exec.Command("git", "config", "--local", "user.email", "robot@mygitops.io"),
	)
	if err != nil {
		t.Fatal("Could not create test git repo in ", repo)
	}

	return repo
}

// Create a new git repo with a `origin` remote useful for push/pull testing
// Returns the path to the new repo
func CreateTestRepoWithOrigin(t *testing.T) (string, string) {
	origin := CreateTestRepo(t, "--bare")
	repo := CreateTestRepo(t)

	envRepo := []string{
		"GIT_DIR=" + path.Join(repo, ".git"),
		"GIT_WORK_TREE=" + repo,
	}

	err := util.RunCommands(envRepo,

		// Add remote and initialize master branch with empty commit
		exec.Command("git", "remote", "add", "origin", origin),
		exec.Command("git", "commit", "--allow-empty", "-m", "Initial Commit (empty)"),
		exec.Command("git", "push", "origin", "master"),
	)

	if err != nil {
		t.Fatal("In repo", repo, ":", err.Error())
	}
	return repo, origin
}

// Create a new git repo seeded with the files in `samplePath`
// Returns the path to the new repo
func CreateTestRepoFromSample(t *testing.T, samplePath string) string {
	if samplePath[len(samplePath)-1] == '/' {
		// XXX: Yeah, safer for the `cp -r` below
		t.Fatal("CreateTestRepoFromSample(): paths must not have a trailing slash")
	}

	samplePath, _ = filepath.Abs(samplePath)
	repo, _ := CreateTestRepoWithOrigin(t)

	envRepo := []string{
		"GIT_DIR=" + path.Join(repo, ".git"),
		"GIT_WORK_TREE=" + repo,
	}
	err := util.RunCommands(envRepo,
		exec.Command("cp", "-r", samplePath+"/.", repo),
		exec.Command("git", "add", "-A"),
		exec.Command("git", "commit", "-m", "Initial commit from sample path: "+samplePath),
		exec.Command("git", "push", "origin", "master"),
	)
	if err != nil {
		t.Fatal(err.Error())
	}

	return repo
}
