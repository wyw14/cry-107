package gas

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type COState struct {
	Identity   model.Identity `json:"identity"`
	PPM        float64        `json:"ppm"`
	High       bool           `json:"high"`
	HighHigh   bool           `json:"high_high"`
	ObservedAt time.Time      `json:"observed_at"`
}

type COMonitor struct {
	mu       sync.RWMutex
	high     float64
	highHigh float64
	state    COState
}

func NewCOMonitor(high, highHigh float64) (*COMonitor, error) {
	if high <= 0 || highHigh <= high {
		return nil, fmt.Errorf("CO thresholds are invalid")
	}
	return &COMonitor{high: high, highHigh: highHigh}, nil
}

func (m *COMonitor) Observe(identity model.Identity, ppm float64) (COState, error) {
	if ppm < 0 || !identity.Valid() {
		return COState{}, fmt.Errorf("CO observation is invalid")
	}
	state := COState{
		Identity: identity, PPM: ppm, High: ppm >= m.high,
		HighHigh: ppm >= m.highHigh, ObservedAt: time.Now().UTC(),
	}
	m.mu.Lock()
	m.state = state
	m.mu.Unlock()
	return state, nil
}

func (m *COMonitor) Snapshot() COState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}
