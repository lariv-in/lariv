package views

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lariv-in/lago/getters"
)

func TestAttachRequestLayerExposesQueryAsGet(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/modal?name=example.modal&other=value", nil)
	rec := httptest.NewRecorder()

	AttachRequestLayer{}.Next(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name, err := getters.Key[string]("$get.name")(r.Context())
		if err != nil {
			t.Fatalf("expected name from $get, got error: %v", err)
		}
		if name != "example.modal" {
			t.Fatalf("expected modal name example.modal, got %q", name)
		}
		if other, err := getters.Key[string]("$get.other")(r.Context()); err != nil || other != "value" {
			t.Fatalf("expected other query param from $get, got %q err=%v", other, err)
		}
		w.WriteHeader(http.StatusOK)
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected ok response, got %d", rec.Code)
	}
}
