package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/valer-cara/mgo/pkg/notification"
	"github.com/valer-cara/mgo/pkg/services"
)

const writeErrorsToClient = true

type Server struct {
	listenAddr string
	gitopsRepo string
	kubeconfig string
	helmHome   string
	dryRun     bool

	// XXX: Maybe not the best way to dependency inject this, but sticking
	// this way for now...
	notifier       notification.Notification
	releaseManager services.ReleaseManager
}

func NewServer(listenAddr, gitopsRepo, helmHome, kubeconfig string, notifier notification.Notification, dryRun bool) *Server {
	return &Server{
		listenAddr: listenAddr,

		notifier: notifier,
		releaseManager: services.NewReleaseManagerBatched(&services.ReleaseManagerBatchedOptions{
			GitopsRepo: gitopsRepo,
			KubeConfig: kubeconfig,
			HelmHome:   helmHome,
			DryRun:     dryRun,
		}),
	}
}

func (s *Server) Serve() error {
	log.Println("Starting ReleaseManager...")
	if err := s.releaseManager.Init(); err != nil {
		return err
	}

	deployHandler := DeployHandler{
		releaseManager: s.releaseManager,
		notification:   s.notifier,
	}

	r := mux.NewRouter()
	r.HandleFunc("/", IndexHandler)
	r.Handle("/deploy", deployHandler).Methods("POST")
	http.Handle("/", r)

	log.Println("Server started")
	return http.ListenAndServe(s.listenAddr, nil)
}

func respondf(w http.ResponseWriter, format string, args ...interface{}) {
	w.Write([]byte(fmt.Sprintf(format, args...)))
}

func handleServerError(err error, status int, r *http.Request, w http.ResponseWriter) {
	log.Errorf("[%s] [status: %d] Error: %v", r.RemoteAddr, status, err)

	w.WriteHeader(status)

	if writeErrorsToClient {
		response, _ := json.MarshalIndent(&apiResponseError{
			Status: "error",
			Error:  err.Error(),
		}, "", "  ")

		w.Write(response)
	}
}

type apiResponseError struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}
