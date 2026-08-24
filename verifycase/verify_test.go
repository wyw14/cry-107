package verifycase

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-107/internal/bypass"
	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/pressure"
)

func TestBypassGateClosesAfterDustConveyorDrain(t *testing.T) {
	identity := model.NewIdentity(8)
	service := bypass.NewService()
	if _, err := service.Start(identity, 35); err != nil {
		t.Fatal(err)
	}
	drainer := conveyor.NewDrainer(10, 20*time.Millisecond)
	gate := pressure.NewGate(true)
	shutdown := bypass.NewShutdown(service, drainer, gate)
	if err := shutdown.Prepare(identity, 100); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
	defer cancel()
	if _, err := shutdown.Run(ctx, identity); err == nil {
		t.Fatal("slow conveyor drain unexpectedly completed")
	}
	if state := gate.Snapshot(); !state.Open {
		t.Fatalf("gate closed before conveyor drain proof: %+v", state)
	}
}
