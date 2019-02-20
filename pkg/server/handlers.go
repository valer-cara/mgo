package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/valer-cara/mgo/pkg/deploy"
	"github.com/valer-cara/mgo/pkg/notification"
	"github.com/valer-cara/mgo/pkg/services"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	respondf(w, "hello world!")
}

type DeployHandler struct {
	releaseManager services.ReleaseManager

	// Notification
	notification notification.Notification

	// Request variables passed in
	formTriggerRepo string
	formImageRepo   string
	formImageTag    string
	formAuthor      string
	formCluster     string
}

func (dh DeployHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// passing as pointer, so we can update the form* fields
	status, err := (&dh).init(r)
	if err != nil {
		handleServerError(err, status, r, w)
		dh.sendNotification(err)
		return
	}

	log.Printf("[%s] New deploy request: %s", r.RemoteAddr, dh)
	dopts := dh.getDeployOptions()
	if err := dh.releaseManager.RequestRelease(dopts); err != nil {
		handleServerError(err, http.StatusInternalServerError, r, w)
		dh.sendNotification(err)
		return
	}

	response, err := json.MarshalIndent(apiResponseDeploy{
		Status: "ok",
		Deploy: dh.getDeployOptions(),
	}, "", "  ")
	if err != nil {
		handleServerError(err, http.StatusInternalServerError, r, w)
		dh.sendNotification(err)
		return
	}

	log.Printf("[%s] Deploy successful!", r.RemoteAddr)
	dh.sendNotification(nil)

	w.Write(response)
}

func (dh *DeployHandler) init(r *http.Request) (int, error) {
	if err := r.ParseForm(); err != nil {
		return http.StatusInternalServerError, err
	}

	dh.formTriggerRepo = r.FormValue("triggerRepo")
	dh.formImageRepo = r.FormValue("imageRepo")
	dh.formImageTag = r.FormValue("imageTag")
	dh.formAuthor = r.FormValue("author")
	dh.formCluster = r.FormValue("cluster")

	// If no error, status will be ignored by caller
	// If error, it's a 400 BadRequest
	return http.StatusBadRequest, dh.ValidateInput()
}

func (dh *DeployHandler) ValidateInput() error {
	if dh.formTriggerRepo == "" {
		return errors.New("missing parameter `triggerRepo`")
	}
	if dh.formImageRepo == "" {
		return errors.New("missing parameter `imageRepo`")
	}
	if dh.formImageTag == "" {
		return errors.New("missing parameter `imageTag`")
	}
	if dh.formAuthor == "" {
		return errors.New("missing parameter `author`")
	}
	if dh.formCluster == "" {
		return errors.New("missing parameter `cluster`")
	}
	return nil
}

func (dh *DeployHandler) getDeployOptions() *deploy.DeployOptions {
	return &deploy.DeployOptions{
		TriggerRepo: dh.formTriggerRepo,
		Author:      dh.formAuthor,
		Cluster:     dh.formCluster,
		Image: deploy.DeployOptionsImage{
			Repository: dh.formImageRepo,
			Tag:        dh.formImageTag,
		},
	}
}

func (dh *DeployHandler) sendNotification(err error) error {
	// If there's no notification service defined don't try to send one
	if dh.notification == nil {
		return nil
	}

	log.Debug("Sending notification")
	errNotif := dh.notification.Deployed(
		dh.formTriggerRepo,
		dh.formImageRepo,
		dh.formImageTag,
		dh.formCluster,
		dh.formAuthor,
		err,
	)
	if err != nil {
		log.Error("Error sending notification, err: %v", err)
	}
	return errNotif
}

//func (dh *DeployHandler) deployCreate() (int, error) {
//	deploySvc := services.NewDeployService(dh.gitopsRepo, dh.getDeployOptions())
//
//	if err := deploySvc.Execute(); err != nil {
//		return http.StatusInternalServerError, errors.New(fmt.Sprintf("Failed deployment %v: %v", deploySvc, err))
//	}
//
//	return http.StatusOK, nil
//}
//
//func (dh DeployHandler) deploySync() (int, error) {
//	syncSvc := services.NewSyncService(dh.gitopsRepo, dh.helmHome, dh.kubeconfig, dh.formCluster, dh.dryRun)
//	if err := syncSvc.Init(); err != nil {
//		return http.StatusInternalServerError, errors.New(fmt.Sprintf("Failed init sync %v: %v", syncSvc, err))
//	}
//
//	if err := syncSvc.Execute(); err != nil {
//		return http.StatusInternalServerError, errors.New(fmt.Sprintf("Failed sync %v: %v", syncSvc, err))
//	}
//
//	return http.StatusOK, nil
//}

func (dh DeployHandler) String() string {
	return fmt.Sprintf("triggerRepo: %s, author: %s, cluster: %s, image: %s:%s",
		dh.formTriggerRepo,
		dh.formAuthor,
		dh.formCluster,
		dh.formImageRepo,
		dh.formImageTag,
	)
}

type apiResponseDeploy struct {
	Status string                `json:"status"`
	Error  string                `json:"error"`
	Deploy *deploy.DeployOptions `json:"deploy"`
}
