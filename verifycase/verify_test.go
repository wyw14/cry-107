package verifycase

import (
	"testing"

	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/cooler"
	"github.com/wyw14/cry-107/internal/crusher"
	"github.com/wyw14/cry-107/internal/model"
)

func TestCrusherWaitsForControlledClinkerBacklogRelease(t *testing.T) {
	identity := model.NewIdentity(12)
	transport := conveyor.NewStateService()
	permit := crusher.NewPermit()
	recovery := cooler.NewRecoveryGraph(transport, permit)
	recovery.Begin(identity, 12)
	state, err := recovery.ApplySpeedProof(identity, 100)
	if err != nil {
		t.Fatal(err)
	}
	if state.Crusher.Allowed || state.Transport.BacklogDrained {
		t.Fatalf("belt speed was treated as material drained: %+v", state)
	}
	state, err = recovery.ReleaseBacklog(identity, 12)
	if err != nil || !state.Crusher.Allowed {
		t.Fatalf("controlled release did not complete recovery: %+v %v", state, err)
	}
}
