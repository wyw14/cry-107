package verifycase

import (
	"math"
	"testing"

	"github.com/wyw14/cry-107/internal/kiln"
	"github.com/wyw14/cry-107/internal/pressure"
)

func TestPressureModeTransferStartsWithoutOutputBump(t *testing.T) {
	controller := pressure.NewController(0.7, 0.12, 46)
	controller.Restore(pressure.ControllerState{Setpoint: -45, Integral: 75, Output: 46, LastPV: -45})
	modes := pressure.NewModeService(controller, 46)
	draft := kiln.NewDraftLoop(modes, -30, 46)
	draft.EnterManual()
	if _, err := draft.ManualMove(46); err != nil {
		t.Fatal(err)
	}
	draft.EnterAuto(-45)
	state, err := draft.Tick()
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(state.Damper-46) > 5 {
		t.Fatalf("first automatic output jumped from current damper: %+v", state)
	}
}
