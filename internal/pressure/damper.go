package pressure

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type DamperState struct {
	Identity       model.Identity `json:"identity"`
	Commanded      float64        `json:"commanded"`
	Accepted       bool           `json:"accepted"`
	Actual         float64        `json:"actual"`
	PositionProven bool           `json:"position_proven"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Damper struct {
	mu    sync.RWMutex
	state DamperState
}

func NewDamper(initial float64) *Damper {
	return &Damper{state: DamperState{Commanded: initial, Actual: initial, PositionProven: true}}
}

func (d *Damper) Command(identity model.Identity, target float64) (DamperState, error) {
	if !identity.Valid() || target < 0 || target > 150 {
		return DamperState{}, fmt.Errorf("damper command is invalid")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state.Identity = identity
	d.state.Commanded = target
	d.state.Accepted = true
	d.state.PositionProven = false
	d.state.UpdatedAt = time.Now().UTC()
	return d.state, nil
}

func (d *Damper) Confirm(identity model.Identity, actual float64) (DamperState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if !d.state.Identity.SameGeneration(identity) {
		return DamperState{}, fmt.Errorf("damper position proof is stale")
	}
	d.state.Actual = actual
	d.state.PositionProven = actual+0.5 >= d.state.Commanded
	d.state.UpdatedAt = time.Now().UTC()
	if !d.state.PositionProven {
		return d.state, fmt.Errorf("damper position %.1f has not reached %.1f", actual, d.state.Commanded)
	}
	return d.state, nil
}

func (d *Damper) Snapshot() DamperState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.state
}
