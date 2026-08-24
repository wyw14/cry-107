package grate

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type FanState struct {
	Identity     model.Identity `json:"identity"`
	CommandedAir float64        `json:"commanded_air"`
	ActualAir    float64        `json:"actual_air"`
	Running      bool           `json:"running"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type Fan struct {
	mu    sync.RWMutex
	state FanState
}

func NewFan() *Fan {
	return &Fan{}
}

func (f *Fan) Command(identity model.Identity, air float64) (FanState, error) {
	if !identity.Valid() || air < 0 {
		return FanState{}, fmt.Errorf("grate fan command is invalid")
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state.Identity = identity
	f.state.CommandedAir = air
	f.state.Running = air > 0
	f.state.UpdatedAt = time.Now().UTC()
	return f.state, nil
}

func (f *Fan) Observe(identity model.Identity, air float64) (FanState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.state.Identity.SameGeneration(identity) {
		return f.state, fmt.Errorf("fan observation belongs to another cooling operation")
	}
	f.state.ActualAir = air
	f.state.UpdatedAt = time.Now().UTC()
	return f.state, nil
}

func (f *Fan) Snapshot() FanState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}
