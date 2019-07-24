package server

import (
	"bytes"
	"encoding/json"
	//"net/http"
	"io/ioutil"
	"net/http/httptest"
	//"net/url"
	"strings"
	"testing"

	"github.com/valer-cara/mgo/pkg/services"
)

func TestDockerhub(t *testing.T) {
	payload, err := json.Marshal(
		DockerhubWebhookPayload{
			PushData: DockerhubWebhookPayload_PushData{
				Tag:      "latest",
				Pusher:   "matthew.bellamy@slashdot.org",
				PushedAt: 123456789,
			},
			Repository: DockerhubWebhookPayload_Repository{
				RepoName: "somerepo",
			},
		})

	queryParams := []string{
		"triggerRepo=foo",
		"cluster=bar",
	}

	req := httptest.NewRequest("POST", "/deploy/dockerhub?"+strings.Join(queryParams, "&"), bytes.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := DockerhubHandler{
		releaseManager: &services.ReleaseManagerMock{},
	}
	handler.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatal("Server response != 200OK")
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]interface{}{}
	err = json.Unmarshal(bodyBytes, &body)
	if err != nil {
		t.Fatal(err)
	}

	if body["status"].(string) != "ok" {
		t.Fatalf("Expected 'ok' response, got '%s'", body["status"].(string))
	}
}
