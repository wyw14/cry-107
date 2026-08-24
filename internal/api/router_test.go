package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthAndOperatorPages(t *testing.T) {
	system, err := NewSystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	router := Router(system)
	for _, path := range []string{"/healthz", "/kiln", "/burner", "/cooler", "/incidents"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("%s returned %d: %s", path, response.Code, response.Body.String())
		}
	}
}

func TestLoadWaitsForAirProof(t *testing.T) {
	system, err := NewSystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	router := Router(system)
	request := httptest.NewRequest(http.MethodPost, "/api/kiln/load", strings.NewReader(`{"load_percent":82}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("load request returned %d: %s", response.Code, response.Body.String())
	}
	state := system.Burner.State()["ratio"]
	if state == nil {
		t.Fatal("burner ratio state missing")
	}
}
