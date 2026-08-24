package kiln

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/pressure"
)

type DraftLoop struct {
	mu       sync.RWMutex
	modes    *pressure.ModeService
	pressure float64
	damper   float64
}

func NewDraftLoop(modes *pressure.ModeService, pressurePV, damper float64) *DraftLoop {
	return &DraftLoop{modes: modes, pressure: pressurePV, damper: damper}
}

func (d *DraftLoop) Observe(value float64) {
	d.mu.Lock()
	d.pressure = value
	d.mu.Unlock()
}

func (d *DraftLoop) Tick() (pressure.ModeState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, err := d.modes.Tick(d.pressure)
	if err != nil {
		return state, fmt.Errorf("draft automatic cycle: %w", err)
	}
	d.damper = state.Damper
	return state, nil
}

func (d *DraftLoop) EnterAuto(setpoint float64) pressure.ModeState {
	d.mu.Lock()
	defer d.mu.Unlock()
	state := d.modes.EnterAuto(setpoint, d.pressure, d.damper)
	d.damper = state.Damper
	return state
}

func (d *DraftLoop) EnterManual() pressure.ModeState {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.modes.EnterManual(d.damper, d.pressure)
}

func (d *DraftLoop) ManualMove(target float64) (pressure.ModeState, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	state, err := d.modes.ManualMove(target, d.pressure)
	if err == nil {
		d.damper = state.Damper
	}
	return state, err
}

func (d *DraftLoop) Snapshot() (float64, float64, pressure.ModeState) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.pressure, d.damper, d.modes.Snapshot()
}
