package burner

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type RatioPlan struct {
	Identity    model.Identity `json:"identity"`
	LoadPercent float64        `json:"load_percent"`
	RequiredAir float64        `json:"required_air"`
	TargetFuel  float64        `json:"target_fuel"`
	CreatedAt   time.Time      `json:"created_at"`
}

type RatioState struct {
	Plan            RatioPlan `json:"plan"`
	AcceptedAir     float64   `json:"accepted_air"`
	ConfirmedAir    float64   `json:"confirmed_air"`
	FuelTarget      float64   `json:"fuel_target"`
	AirProofPending bool      `json:"air_proof_pending"`
	Failure         string    `json:"failure,omitempty"`
}

type RatioLoop struct {
	mu    sync.RWMutex
	state RatioState
}

func NewRatioLoop(initialFuel, initialAir float64) *RatioLoop {
	return &RatioLoop{state: RatioState{FuelTarget: initialFuel, ConfirmedAir: initialAir}}
}

func (l *RatioLoop) Propose(identity model.Identity, loadPercent float64) (RatioPlan, error) {
	if !identity.Valid() || loadPercent < 20 || loadPercent > 100 {
		return RatioPlan{}, fmt.Errorf("burner load plan is outside operating envelope")
	}
	plan := RatioPlan{
		Identity: identity, LoadPercent: loadPercent,
		RequiredAir: loadPercent * 1.25, TargetFuel: loadPercent * 0.22,
		CreatedAt: time.Now().UTC(),
	}
	l.mu.Lock()
	l.state.Plan = plan
	l.state.AirProofPending = true
	l.state.Failure = ""
	l.mu.Unlock()
	return plan, nil
}

func (l *RatioLoop) AirCommandAccepted(identity model.Identity, target float64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.state.Plan.Identity.SameGeneration(identity) {
		return fmt.Errorf("air command acknowledgement is stale")
	}
	// The damper has accepted the move command, but combustion air is not yet
	// established — the log shows "air command accepted" landing before "damper
	// position reached". Only record the accepted target here; leave AirProofPending
	// raised and do not touch ConfirmedAir or FuelTarget. Fuel advances solely in
	// AirConfirmed once real airflow is proven, so coal can never ramp into an
	// air-starved kiln and drive the flame reducing.
	l.state.AcceptedAir = target
	return nil
}

// airProofTolerance mirrors the combustion-air damper's position-reached
// band (see pressure.Damper.Confirm: actual+0.5 >= Commanded). The fuel gate
// must use the same band so that, whenever the damper regards the air move as
// established, coal is allowed to follow — a normally responding damper must
// not be held back by a stricter gate. Air that falls short of the damper's own
// proof still blocks fuel here, preserving the air-before-fuel interlock.
const airProofTolerance = 0.5

func (l *RatioLoop) AirConfirmed(identity model.Identity, actual float64) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.state.Plan.Identity.SameGeneration(identity) {
		return fmt.Errorf("air proof belongs to another control cycle")
	}
	if actual+airProofTolerance < l.state.Plan.RequiredAir {
		l.state.Failure = fmt.Sprintf("combustion air %.1f below required %.1f", actual, l.state.Plan.RequiredAir)
		return fmt.Errorf("%s", l.state.Failure)
	}
	l.state.ConfirmedAir = actual
	l.state.FuelTarget = l.state.Plan.TargetFuel
	l.state.AirProofPending = false
	return nil
}

func (l *RatioLoop) FailAir(identity model.Identity, reason string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.state.Plan.Identity.SameGeneration(identity) {
		return fmt.Errorf("air failure belongs to another control cycle")
	}
	l.state.Failure = reason
	l.state.AirProofPending = false
	return nil
}

func (l *RatioLoop) Snapshot() RatioState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}
