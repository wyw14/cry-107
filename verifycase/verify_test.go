package verifycase

import (
	"testing"

	"github.com/wyw14/cry-107/internal/bearing"
	"github.com/wyw14/cry-107/internal/kiln"
	"github.com/wyw14/cry-107/internal/lube"
	"github.com/wyw14/cry-107/internal/model"
)

func TestThrustRollerResumeRequiresReturnOilFlow(t *testing.T) {
	identity := model.NewIdentity(4)
	flow := bearing.NewFlowObserver(18)
	handover := lube.NewCoordinator(flow, 2.4)
	if _, err := handover.Begin(identity, "standby"); err != nil {
		t.Fatal(err)
	}
	status, err := handover.ConfirmPressure(identity, 3.2)
	if err != nil {
		t.Fatal(err)
	}
	if status.Complete {
		t.Fatalf("pressure alone completed handover: %+v", status)
	}
	permit := kiln.NewDrivePermit()
	permit.ApplyHandover(status)
	if err := permit.Check(identity, 2.2); err == nil {
		t.Fatal("kiln drive resumed without return-oil proof")
	}
	status, err = handover.ConfirmReturnFlow(identity, 22)
	if err != nil || !status.Complete {
		t.Fatalf("matching return-flow proof did not complete handover: %+v %v", status, err)
	}
}
