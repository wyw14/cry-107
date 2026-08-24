package coalmill

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/gas"
)

// recordingPermit records every inerting transition so tests can assert that
// hot-air and coal-feed permission only persists while the oxygen window is
// genuinely proven, and is revoked the moment the window breaks.
type recordingPermit struct {
	started   bool
	completed bool
	hotAir    bool
	coalFeed  bool
	reasons   []string
}

func (p *recordingPermit) InertingStarted(uuid.UUID) {
	p.started = true
	p.hotAir = false
	p.coalFeed = false
}

func (p *recordingPermit) InertingCompleted(uuid.UUID) bool {
	p.completed = true
	p.hotAir = true
	p.coalFeed = true
	p.reasons = append(p.reasons, "completed")
	return true
}

func (p *recordingPermit) InertingRevoked(_ uuid.UUID, reason string) {
	p.hotAir = false
	p.coalFeed = false
	p.reasons = append(p.reasons, reason)
}

func newProdInerting() (*InertingService, *recordingPermit) {
	// Production configuration: 8% oxygen held for 30 s, 900 m3 purge volume.
	window := gas.NewOxygenWindow(8, 30*time.Second)
	permit := &recordingPermit{}
	return NewInertingService(window, permit, 900), permit
}

// TestInertingNeedsVolumeAndWindow proves that purge volume alone or a single
// instantaneous low-oxygen frame can never mark inerting complete.
func TestInertingNeedsVolumeAndWindow(t *testing.T) {
	svc, permit := newProdInerting()
	st := svc.Begin()
	// Volume alone must not complete inerting.
	if state, _ := svc.AddNitrogen(st.SessionID, 950); state.Complete || permit.hotAir {
		t.Fatalf("purge volume alone completed inerting: Complete=%v hotAir=%v", state.Complete, permit.hotAir)
	}
	start := time.Unix(1000, 0)
	// High oxygen frame then a SINGLE low frame at the limit: StableFor == 0.
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 13, Observed: start})
	if state, _ := svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 8, Observed: start.Add(10 * time.Second)}); state.Complete || permit.hotAir {
		t.Fatalf("single instantaneous low frame completed inerting: Complete=%v StableFor=%v", state.Complete, state.Oxygen.StableFor)
	}
}

// TestInertingCompletesOnContinuousWindow mirrors the bootstrap sequence: with
// sufficient purge volume, two low frames 31 s apart must complete inerting.
func TestInertingCompletesOnContinuousWindow(t *testing.T) {
	svc, permit := newProdInerting()
	st := svc.Begin()
	svc.AddNitrogen(st.SessionID, 950)
	start := time.Unix(1000, 0)
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.6, Observed: start})
	if state, _ := svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.4, Observed: start.Add(31 * time.Second)}); !state.Complete || !permit.hotAir {
		t.Fatalf("continuous 30 s low-oxygen window did not complete inerting: Complete=%v hotAir=%v", state.Complete, permit.hotAir)
	}
}

// TestInertingRevokedOnOxygenSurge is the regression for the reported incident:
// once inerting completes, the very next frame surging above 8% must revoke the
// hot-air permit so the damper closes before a high-oxygen interlock trips.
func TestInertingRevokedOnOxygenSurge(t *testing.T) {
	svc, permit := newProdInerting()
	st := svc.Begin()
	svc.AddNitrogen(st.SessionID, 950)
	start := time.Unix(1000, 0)
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.6, Observed: start})
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.4, Observed: start.Add(31 * time.Second)})
	if !permit.hotAir {
		t.Fatal("precondition: hot-air permit should be open after continuous window")
	}
	// Next frame surges back to 13% as in the field incident.
	state, _ := svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 13, Observed: start.Add(32 * time.Second)})
	if state.Complete || permit.hotAir {
		t.Fatalf("oxygen surge above limit did not revoke inerting: Complete=%v hotAir=%v", state.Complete, permit.hotAir)
	}
}

// TestInertingRevokedWhenVolumeIsInsufficient ensures the permit cannot persist
// if purge volume drops below the required floor while oxygen stays low.
func TestInertingRevokedWhenVolumeIsInsufficient(t *testing.T) {
	svc, permit := newProdInerting()
	st := svc.Begin()
	start := time.Unix(1000, 0)
	// Build a proven low-oxygen window first.
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.6, Observed: start})
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.4, Observed: start.Add(31 * time.Second)})
	// Volume was never added: completion must stay false.
	if state := svc.Snapshot(); state.Complete {
		t.Fatalf("inerting completed without required purge volume: Complete=%v", state.Complete)
	}
	if permit.hotAir {
		t.Fatalf("hot-air permit opened without required purge volume")
	}
}

// TestInertingReCompletesAfterRecovery ensures the state machine can grant the
// permit again once oxygen returns below the limit and the window is re-proven,
// proving the revoke is non-terminal (unlike a CO trip).
func TestInertingReCompletesAfterRecovery(t *testing.T) {
	svc, permit := newProdInerting()
	st := svc.Begin()
	svc.AddNitrogen(st.SessionID, 950)
	start := time.Unix(1000, 0)
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.6, Observed: start})
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.4, Observed: start.Add(31 * time.Second)})
	// Surge above the limit, breaking the window and revoking the permit.
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 13, Observed: start.Add(32 * time.Second)})
	if permit.hotAir {
		t.Fatal("hot-air permit should be revoked after surge")
	}
	// Recover: oxygen drops back and a fresh 30 s window is re-proven.
	svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.5, Observed: start.Add(40 * time.Second)})
	if state, _ := svc.ObserveOxygen(gas.OxygenSample{SessionID: st.SessionID, Percent: 7.3, Observed: start.Add(71 * time.Second)}); !state.Complete || !permit.hotAir {
		t.Fatalf("inerting did not re-complete after recovery: Complete=%v hotAir=%v", state.Complete, permit.hotAir)
	}
}
