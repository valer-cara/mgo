package server

import (
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"

	"github.com/valer-cara/mgo/pkg/deploy"
	"github.com/valer-cara/mgo/pkg/notification"
	"github.com/valer-cara/mgo/pkg/services"
)

// Reference here:
// https://docs.docker.com/docker-hub/webhooks/#example-webhook-payload
// https://mholt.github.io/json-to-go/
type DockerhubWebhookPayload struct {
	CallbackURL string                             `json:"callback_url"`
	PushData    DockerhubWebhookPayload_PushData   `json:"push_data"`
	Repository  DockerhubWebhookPayload_Repository `json:"repository"`
}
type DockerhubWebhookPayload_PushData struct {
	Images   []string `json:"images"`
	PushedAt int      `json:"pushed_at"`
	Pusher   string   `json:"pusher"`
	Tag      string   `json:"tag"`
}
type DockerhubWebhookPayload_Repository struct {
	CommentCount    int    `json:"comment_count"`
	DateCreated     int    `json:"date_created"`
	Description     string `json:"description"`
	Dockerfile      string `json:"dockerfile"`
	FullDescription string `json:"full_description"`
	IsOfficial      bool   `json:"is_official"`
	IsPrivate       bool   `json:"is_private"`
	IsTrusted       bool   `json:"is_trusted"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Owner           string `json:"owner"`
	RepoName        string `json:"repo_name"`
	RepoURL         string `json:"repo_url"`
	StarCount       int    `json:"star_count"`
	Status          string `json:"status"`
}

type DockerhubHandler struct {
	releaseManager services.ReleaseManager
	notification   notification.Notification

	// Filled in from request
	payload     *DockerhubWebhookPayload
	triggerRepo string
	cluster     string
}

func (dh DockerhubHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

func (dh *DockerhubHandler) init(r *http.Request) (int, error) {
	var payload DockerhubWebhookPayload

	if err := r.ParseForm(); err != nil {
		return http.StatusInternalServerError, err
	}

	dh.triggerRepo = r.FormValue("triggerRepo")
	dh.cluster = r.FormValue("cluster")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	dh.payload = &payload

	// If no error, status will be ignored by caller
	// If error, it's a 400 BadRequest
	return http.StatusBadRequest, dh.ValidateInput()
}

func (dh *DockerhubHandler) ValidateInput() error {
	if dh.triggerRepo == "" {
		return errors.New("missing http parameter `triggerRepo`")
	}
	if dh.cluster == "" {
		return errors.New("missing http parameter `cluster`")
	}
	if dh.payload.Repository.RepoName == "" {
		return errors.New("is json payload from dockerhub malformed? No `repository_name` found in request")
	}
	if dh.payload.PushData.Pusher == "" {
		return errors.New("is json payload from dockerhub malformed? No `pusher` found in request")
	}
	if dh.payload.PushData.Tag == "" {
		return errors.New("is json payload from dockerhub malformed? No `tag` found in request")
	}
	return nil
}

func (dh *DockerhubHandler) getDeployOptions() *deploy.DeployOptions {
	return &deploy.DeployOptions{
		TriggerRepo: dh.triggerRepo,
		Author:      dh.payload.PushData.Pusher,
		Cluster:     dh.cluster,
		Image: deploy.DeployOptionsImage{
			Repository: dh.payload.Repository.RepoName,
			Tag:        dh.payload.PushData.Tag,
		},
	}
}

// XXX: terrible dupes.. use a middleware for errors & notifs...
func (dh *DockerhubHandler) sendNotification(err error) error {
	// If there's no notification service defined don't try to send one
	if dh.notification == nil {
		return nil
	}

	log.Debug("Sending notification")
	errNotif := dh.notification.Deployed(
		dh.triggerRepo,
		dh.payload.Repository.Name,
		dh.payload.PushData.Tag,
		dh.cluster,
		dh.payload.PushData.Pusher,
		err,
	)
	if err != nil {
		log.Errorf("Error sending notification, err: %v", err)
	}
	return errNotif
}

func (dh DockerhubHandler) String() string {
	return fmt.Sprintf("triggerRepo: %s, author: %s, cluster: %s, image: %s:%s",
		dh.triggerRepo,
		dh.payload.PushData.Pusher,
		dh.cluster,
		dh.payload.Repository.RepoName,
		dh.payload.PushData.Tag,
	)
}
