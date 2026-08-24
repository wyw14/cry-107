package bearing

import (
	"fmt"
	"sync"
	"time"
)

type TemperatureState struct {
	Celsius    float64   `json:"celsius"`
	Alarm      bool      `json:"alarm"`
	Trip       bool      `json:"trip"`
	ObservedAt time.Time `json:"observed_at"`
}

type TemperatureMonitor struct {
	mu    sync.RWMutex
	alarm float64
	trip  float64
	state TemperatureState
}

func NewTemperatureMonitor(alarm, trip float64) (*TemperatureMonitor, error) {
	if alarm <= 0 || trip <= alarm {
		return nil, fmt.Errorf("bearing temperature limits are invalid")
	}
	return &TemperatureMonitor{alarm: alarm, trip: trip}, nil
}

func (m *TemperatureMonitor) Observe(value float64) TemperatureState {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = TemperatureState{
		Celsius: value, Alarm: value >= m.alarm, Trip: value >= m.trip, ObservedAt: time.Now().UTC(),
	}
	return m.state
}

func (m *TemperatureMonitor) Snapshot() TemperatureState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}
