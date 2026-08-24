package gas

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type OxygenSample struct {
	SessionID uuid.UUID `json:"session_id"`
	Percent   float64   `json:"percent"`
	Observed  time.Time `json:"observed_at"`
}

type OxygenState struct {
	SessionID    uuid.UUID     `json:"session_id"`
	Latest       float64       `json:"latest_percent"`
	StableFor    time.Duration `json:"stable_for"`
	Stable       bool          `json:"stable"`
	SampleCount  int           `json:"sample_count"`
	LastObserved time.Time     `json:"last_observed_at"`
}

type OxygenWindow struct {
	mu       sync.RWMutex
	limit    float64
	required time.Duration
	state    OxygenState
	lowSince time.Time
}

func NewOxygenWindow(limit float64, required time.Duration) *OxygenWindow {
	return &OxygenWindow{limit: limit, required: required}
}

func (w *OxygenWindow) Reset(session uuid.UUID) {
	w.mu.Lock()
	w.state = OxygenState{SessionID: session}
	w.lowSince = time.Time{}
	w.mu.Unlock()
}

func (w *OxygenWindow) Add(sample OxygenSample) (OxygenState, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if sample.SessionID == uuid.Nil || sample.SessionID != w.state.SessionID {
		return w.state, fmt.Errorf("oxygen sample belongs to another inerting session")
	}
	if sample.Percent < 0 || sample.Percent > 25 || sample.Observed.IsZero() {
		return w.state, fmt.Errorf("oxygen sample is invalid")
	}
	if !w.state.LastObserved.IsZero() && sample.Observed.Before(w.state.LastObserved) {
		return w.state, fmt.Errorf("oxygen sample arrived out of order")
	}
	w.state.Latest = sample.Percent
	w.state.SampleCount++
	w.state.LastObserved = sample.Observed
	if sample.Percent <= w.limit {
		if w.lowSince.IsZero() {
			w.lowSince = sample.Observed
		}
		w.state.StableFor = sample.Observed.Sub(w.lowSince)
		w.state.Stable = w.state.StableFor >= w.required
	} else {
		w.lowSince = time.Time{}
		w.state.StableFor = 0
		w.state.Stable = false
	}
	return w.state, nil
}

func (w *OxygenWindow) Snapshot() OxygenState {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.state
}
