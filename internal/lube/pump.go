package lube

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type PumpState struct {
	Identity model.Identity `json:"identity"`
	Name     string         `json:"name"`
	Running  bool           `json:"running"`
	Pressure float64        `json:"pressure_bar"`
	Changed  time.Time      `json:"changed_at"`
}

type PumpBank struct {
	mu    sync.RWMutex
	pumps map[string]PumpState
}

func NewPumpBank(names ...string) *PumpBank {
	bank := &PumpBank{pumps: make(map[string]PumpState)}
	for _, name := range names {
		bank.pumps[name] = PumpState{Name: name}
	}
	return bank
}

func (b *PumpBank) Start(identity model.Identity, name string) (PumpState, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	state, ok := b.pumps[name]
	if !ok {
		return PumpState{}, fmt.Errorf("unknown lubrication pump %s", name)
	}
	state.Identity = identity
	state.Running = true
	state.Changed = time.Now().UTC()
	b.pumps[name] = state
	return state, nil
}

func (b *PumpBank) ObservePressure(identity model.Identity, name string, pressure float64) (PumpState, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	state, ok := b.pumps[name]
	if !ok || !state.Identity.SameGeneration(identity) {
		return PumpState{}, fmt.Errorf("pump pressure proof does not match active pump")
	}
	state.Pressure = pressure
	state.Changed = time.Now().UTC()
	b.pumps[name] = state
	return state, nil
}

func (b *PumpBank) List() []PumpState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	values := make([]PumpState, 0, len(b.pumps))
	for _, state := range b.pumps {
		values = append(values, state)
	}
	return values
}
