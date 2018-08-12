package server

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/valer-cara/mgo/pkg/services"
)

func TestServerIndex(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	IndexHandler(w, req)

	resp := w.Result()
	if resp.StatusCode != 200 {
		t.Fatal("Server response != 200OK")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	expected := "hello world!"
	if string(body) != expected {
		t.Fatalf("Expected '%s' response, got '%s'", expected, string(body))
	}
}

func TestServerDeployHandlerValidation(t *testing.T) {
	for _, removedKey := range []string{"triggerRepo", "imageRepo", "imageTag", "author", "cluster"} {
		data := url.Values{}
		data.Set("triggerRepo", "xxx")
		data.Set("imageRepo", "xxx")
		data.Set("imageTag", "xxx")
		data.Set("author", "xxx")
		data.Set("cluster", "xxx")

		data.Del(removedKey)

		req := httptest.NewRequest("POST", "/deploy", strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()

		handler := DeployHandler{}
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected status 400 BadRequest with missing param %s. Got %d.", removedKey, resp.StatusCode)
		}
	}
}

func TestServerDeployHandlerResponses(t *testing.T) {
	data := url.Values{}
	data.Set("triggerRepo", "xxx")
	data.Set("imageRepo", "xxx")
	data.Set("imageTag", "xxx")
	data.Set("author", "xxx")
	data.Set("cluster", "xxx")

	tests := []struct {
		ReleaseManager services.ReleaseManager
		Status         int
	}{
		{&services.ReleaseManagerMock{}, http.StatusOK},
		{&services.ReleaseManagerMock{RequestReleaseError: errors.New("request_release")}, http.StatusInternalServerError},
	}

	for testIdx, test := range tests {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/deploy", strings.NewReader(data.Encode()))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		handler := DeployHandler{
			releaseManager: test.ReleaseManager,
		}
		handler.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != test.Status {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Fatalf("[test %d] Expected status %d, got %d. Response was:\n %s", testIdx, test.Status, resp.StatusCode, string(b))
		}
	}
}
