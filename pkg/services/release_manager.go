package services

import (
	"github.com/valer-cara/mgo/pkg/deploy"
)

// ReleaseManager: Aggregates multiple requests to deploy, does each
// edit/commit as requested and does a cluster sync at the end
// Should be global
type ReleaseManager interface {
	Init() error
	RequestRelease(*deploy.DeployOptions) error
}
