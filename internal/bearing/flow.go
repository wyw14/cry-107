package bearing

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type FlowState struct {
	Identity   model.Identity `json:"identity"`
	LitresMin  float64        `json:"litres_min"`
	Proven     bool           `json:"proven"`
	ObservedAt time.Time      `json:"observed_at"`
}

type FlowObserver struct {
	mu      sync.RWMutex
	minimum float64
	state   FlowState
}

func NewFlowObserver(minimum float64) *FlowObserver {
	return &FlowObserver{minimum: minimum}
}

func (o *FlowObserver) Begin(identity model.Identity) {
	o.mu.Lock()
	o.state = FlowState{Identity: identity, ObservedAt: time.Now().UTC()}
	o.mu.Unlock()
}

func (o *FlowObserver) Observe(identity model.Identity, flow float64) (FlowState, error) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if !o.state.Identity.SameGeneration(identity) {
		return o.state, fmt.Errorf("return-oil observation belongs to another pump handover")
	}
	o.state.LitresMin = flow
	o.state.Proven = flow >= o.minimum
	o.state.ObservedAt = time.Now().UTC()
	return o.state, nil
}

func (o *FlowObserver) Snapshot() FlowState {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.state
}
