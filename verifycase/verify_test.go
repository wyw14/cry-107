package verifycase

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/wyw14/cry-107/internal/api"
	"github.com/wyw14/cry-107/internal/burner"
)

func TestFuelIncreaseWaitsForCombustionAirProof(t *testing.T) {
	system, err := api.NewSystem(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	router := api.Router(system)
	request := httptest.NewRequest(http.MethodPost, "/api/kiln/load", strings.NewReader(`{"load_percent":82}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("load request returned %d: %s", response.Code, response.Body.String())
	}
	before := system.Burner.State()["ratio"].(burner.RatioState)
	if before.FuelTarget != 15.4 || !before.AirProofPending {
		t.Fatalf("fuel advanced before physical air proof: %+v", before)
	}
	proof := httptest.NewRequest(http.MethodPost, "/api/kiln/air-proof", strings.NewReader(`{"actual_air":103}`))
	proof.Header.Set("Content-Type", "application/json")
	confirmed := httptest.NewRecorder()
	router.ServeHTTP(confirmed, proof)
	if confirmed.Code != http.StatusOK {
		t.Fatalf("air proof returned %d: %s", confirmed.Code, confirmed.Body.String())
	}
	after := system.Burner.State()["ratio"].(burner.RatioState)
	if after.FuelTarget <= before.FuelTarget || after.AirProofPending {
		t.Fatalf("fuel did not advance after matching proof: %+v", after)
	}
}
