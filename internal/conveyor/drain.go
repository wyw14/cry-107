package conveyor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type DrainState struct {
	Identity  model.Identity `json:"identity"`
	Remaining float64        `json:"remaining_kg"`
	Draining  bool           `json:"draining"`
	Drained   bool           `json:"drained"`
	Failure   string         `json:"failure,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`
}

type Drainer struct {
	mu       sync.RWMutex
	rate     float64
	interval time.Duration
	state    DrainState
}

func NewDrainer(rate float64, interval time.Duration) *Drainer {
	return &Drainer{rate: rate, interval: interval}
}

func (d *Drainer) Load(identity model.Identity, kilograms float64) {
	d.mu.Lock()
	d.state = DrainState{Identity: identity, Remaining: kilograms, UpdatedAt: time.Now().UTC()}
	d.mu.Unlock()
}

func (d *Drainer) Drain(ctx context.Context, identity model.Identity) (DrainState, error) {
	d.mu.Lock()
	if !d.state.Identity.SameGeneration(identity) {
		d.mu.Unlock()
		return DrainState{}, fmt.Errorf("drain request belongs to another bypass operation")
	}
	d.state.Draining = true
	d.state.Failure = ""
	d.mu.Unlock()
	ticker := time.NewTicker(d.interval)
	defer ticker.Stop()
	for {
		d.mu.Lock()
		if d.state.Remaining <= 0 {
			d.state.Remaining = 0
			d.state.Draining = false
			d.state.Drained = true
			d.state.UpdatedAt = time.Now().UTC()
			state := d.state
			d.mu.Unlock()
			return state, nil
		}
		d.mu.Unlock()
		select {
		case <-ctx.Done():
			d.mu.Lock()
			d.state.Draining = false
			d.state.Failure = "dust conveyor drain canceled"
			d.state.UpdatedAt = time.Now().UTC()
			state := d.state
			d.mu.Unlock()
			return state, fmt.Errorf("dust conveyor drain canceled: %w", ctx.Err())
		case <-ticker.C:
			d.mu.Lock()
			d.state.Remaining -= d.rate
			d.state.UpdatedAt = time.Now().UTC()
			d.mu.Unlock()
		}
	}
}

func (d *Drainer) Snapshot() DrainState {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.state
}
