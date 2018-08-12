package git

import (
	"errors"
	//log "github.com/sirupsen/logrus"
)

type Git struct {
	backend GitBackend
}

type GitBackend interface {
	Init(string) error
	Pull(extraArgs ...string) error
	Reset() error
	Fetch() error
	Push() error
	AddAll() error
	Commit(string) error

	Root() string
}

const (
	BACKEND_EXTERNAL = 1
	BACKEND_FAKE     = 999
)

func NewGit(backend int, path string) (*Git, error) {
	var bk GitBackend

	switch backend {
	case BACKEND_EXTERNAL:
		bk = NewGitBackendExternal()

		err := bk.Init(path)
		if err != nil {
			return nil, err
		}

	case BACKEND_FAKE:
		bk = &FakeGitBackend{}
		err := bk.Init("ignored")
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("NewGit(): no such backend " + string(backend))
	}

	return &Git{
		backend: bk,
	}, nil
}

func (g *Git) Root() string {
	return g.backend.Root()
}
func (g *Git) Fetch() error {
	return g.backend.Fetch()
}
func (g *Git) Reset() error {
	return g.backend.Reset()
}
func (g *Git) Pull(extraArgs ...string) error {
	return g.backend.Pull(extraArgs...)
}
func (g *Git) Push() error {
	return g.backend.Push()
}
func (g *Git) AddAll() error {
	return g.backend.AddAll()
}
func (g *Git) Commit(msg string) error {
	return g.backend.Commit(msg)
}
