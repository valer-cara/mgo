package git

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	//log "github.com/sirupsen/logrus"
	"os/exec"
	"path"
)

const (
	GIT_EMAIL = "mygitops@foo.bar"
	GIT_NAME  = "MyGitops robot"
)

type GitBackendExternal struct {
	repoPath string
	branch   string
}

func NewGitBackendExternal() *GitBackendExternal {
	return &GitBackendExternal{
		branch: "master",
	}
}

func (g *GitBackendExternal) Init(repoPath string) error {
	if repoPath == "" {
		return errors.New("No path provided for repository")
	}
	g.repoPath = repoPath

	var out bytes.Buffer

	cmd := g.craftGitCommand("rev-parse", "--git-dir")
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New(out.String())
	}

	return nil
}

func (g *GitBackendExternal) Root() string {
	return g.repoPath
}

func (g *GitBackendExternal) Reset() error {
	var out bytes.Buffer

	cmd := g.craftGitCommand("reset", "--hard", "origin/"+g.branch)
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.Reset(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) Fetch() error {
	var out bytes.Buffer

	cmd := g.craftGitCommand("fetch")
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.Fetch(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) Pull(extraArgs ...string) error {
	var out bytes.Buffer

	args := append([]string{"pull", "origin", g.branch}, extraArgs...)

	cmd := g.craftGitCommand(args...)
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.Pull(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) Push() error {
	var out bytes.Buffer

	cmd := g.craftGitCommand("push", "origin", g.branch)
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.Push(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) AddAll() error {
	var out bytes.Buffer

	cmd := g.craftGitCommand("add", "-A")
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.AddAll(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) Commit(msg string) error {
	var out bytes.Buffer

	cmd := g.craftGitCommand("commit", "--allow-empty", "-m", msg)
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return errors.New("Git.Commit(): " + out.String())
	}
	return nil
}

func (g *GitBackendExternal) craftGitCommand(extraArgs ...string) *exec.Cmd {
	args := append([]string{
		"-c", "user.name='" + GIT_NAME + "'",
		"-c", "user.email='" + GIT_EMAIL + "'",
	}, extraArgs...)

	cmd := exec.Command("git", args...)
	log.Debugln("  - running: git", args)
	cmd.Env = []string{
		"GIT_DIR=" + path.Join(g.repoPath, ".git"),
		"GIT_WORK_TREE=" + g.repoPath,
	}

	return cmd
}
