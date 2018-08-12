package helm

import (
	"io"
)

// Abstraction providing helm services
type HelmService interface {
	Init() error
	SyncRelease(*HelmRelease, []string) error
	DiffRelease(*HelmRelease, []string) error
	AddRepo(*HelmRepo) error
	ListRepos() ([]HelmRepo, error)
	UpdateRepos() error

	SetOutput(io.Writer)
}

// Structure containing info for a helm release
type HelmRelease struct {
	// Chart name, in the form of "repo/name". Eg: stable/redis
	Chart string

	// Chart version. As seen in the `Chart.yaml` file of the chart.
	Version string

	// Release name, the name you installed this release as
	Name string

	// Release namespace
	Namespace string
}

// Structure containing info for a helm repository
type HelmRepo struct {
	Name string
	Url  string
}
