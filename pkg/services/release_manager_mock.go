package services

import (
	"github.com/valer-cara/mgo/pkg/deploy"
	log "github.com/sirupsen/logrus"
)

type ReleaseManagerMock struct {
	InitError           error
	RequestReleaseError error
}

func (r *ReleaseManagerMock) Init() error {
	log.Println("ReleaseManagerMock: Init()")
	if r.InitError != nil {
		return r.InitError
	}
	return nil
}

func (r *ReleaseManagerMock) RequestRelease(dopts *deploy.DeployOptions) error {
	log.Println("ReleaseManagerMock: RequestRelease()")
	if r.RequestReleaseError != nil {
		return r.RequestReleaseError
	}
	return nil
}
