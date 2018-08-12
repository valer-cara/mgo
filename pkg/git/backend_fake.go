package git

import (
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

type FakeGitBackend struct {
	repoPath string
}

func (g *FakeGitBackend) Init(ignored string) error {
	log.Println("FakeGit: Init")
	dir, err := ioutil.TempDir("/tmp/", "_mygitops-fake-git-backend-")
	if err != nil {
		return err
	}

	g.repoPath = dir

	return nil
}
func (g *FakeGitBackend) Root() string {
	log.Println("FakeGit: Root")
	return g.repoPath
}
func (g *FakeGitBackend) Fetch() error {
	log.Println("FakeGit: Fetch")
	return nil
}
func (g *FakeGitBackend) Reset() error {
	log.Println("FakeGit: Reset")
	return nil
}
func (g *FakeGitBackend) Pull(extraArgs ...string) error {
	log.Println("FakeGit: Pull", extraArgs)
	return nil
}
func (g *FakeGitBackend) Push() error {
	log.Println("FakeGit: Push")
	return nil
}
func (g *FakeGitBackend) AddAll() error {
	log.Println("FakeGit: AddAll")
	return nil
}
func (g *FakeGitBackend) Commit(msg string) error {
	log.Printf("FakeGit: Commit with message \"%s\"\n", msg)
	return nil
}
