package helm

import (
	"errors"
	"io"
)

// Mock Helm Service
// Set the `FailOn*` values to return an error with that message. If not
// set/empty, the corresponding calls will succeseed
type HelmFake struct {
	FailOnInit        string
	FailOnSyncRelease string
	FailOnDiffRelease string
	FailOnAddRepo     string
	FailOnListRepos   string
	FailOnUpdateRepos string
	Repos             []HelmRepo
}

func (h *HelmFake) Init() error {
	if h.FailOnInit != "" {
		return errors.New(h.FailOnInit)
	}
	return nil
}
func (h *HelmFake) SyncRelease(*HelmRelease, []string) error {
	if h.FailOnSyncRelease != "" {
		return errors.New(h.FailOnSyncRelease)
	}
	return nil
}
func (h *HelmFake) DiffRelease(*HelmRelease, []string) error {
	if h.FailOnDiffRelease != "" {
		return errors.New(h.FailOnDiffRelease)
	}
	return nil
}
func (h *HelmFake) AddRepo(*HelmRepo) error {
	if h.FailOnInit != "" {
		return errors.New(h.FailOnInit)
	}
	return nil
}
func (h *HelmFake) ListRepos() ([]HelmRepo, error) {
	if h.FailOnListRepos != "" {
		return nil, errors.New(h.FailOnListRepos)
	}
	return h.Repos, nil
}
func (h *HelmFake) UpdateRepos() error {
	if h.FailOnUpdateRepos != "" {
		return errors.New(h.FailOnUpdateRepos)
	}
	return nil
}
func (h *HelmFake) SetOutput(io.Writer) {
}
