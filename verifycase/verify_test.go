package verifycase

import (
	"testing"
	"time"

	"github.com/wyw14/cry-107/internal/coalmill"
	"github.com/wyw14/cry-107/internal/gas"
	"github.com/wyw14/cry-107/internal/interlock"
)

func TestCoalMillInertingRequiresVolumeAndStableOxygen(t *testing.T) {
	window := gas.NewOxygenWindow(8, 30*time.Second)
	permit := interlock.NewMillPermit()
	service := coalmill.NewInertingService(window, permit, 900)
	state := service.Begin()
	if _, err := service.AddNitrogen(state.SessionID, 950); err != nil {
		t.Fatal(err)
	}
	state, err := service.ObserveOxygen(gas.OxygenSample{SessionID: state.SessionID, Percent: 7.8, Observed: time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	if state.Complete || permit.Snapshot().HotAir || permit.Snapshot().CoalFeed {
		t.Fatalf("one low-oxygen frame completed inerting: %+v %+v", state, permit.Snapshot())
	}
}
