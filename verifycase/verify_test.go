package verifycase

import (
	"context"
	"testing"

	"github.com/wyw14/cry-107/internal/coalmill"
	"github.com/wyw14/cry-107/internal/interlock"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/operation"
)

func TestCOTripBlocksStartupHotAirCompensation(t *testing.T) {
	identity := model.NewIdentity(15)
	manager := interlock.NewManager()
	permit := interlock.NewMillPermit()
	compensator := operation.NewCompensator()
	trip := interlock.NewCOTrip(manager, permit, compensator)
	startup := coalmill.NewStartup(compensator, trip)
	if err := startup.Begin(identity, 35); err != nil {
		t.Fatal(err)
	}
	startup.AdvanceHotAir(18)
	state, err := startup.TripCO(context.Background(), 1800)
	if err != nil {
		t.Fatal(err)
	}
	if !state.TripActive || state.HotAirOpen != 0 || state.CoalFeed {
		t.Fatalf("startup compensation crossed CO terminal: %+v", state)
	}
}
