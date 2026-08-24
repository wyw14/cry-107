package lube

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/bearing"
	"github.com/wyw14/cry-107/internal/model"
)

type HandoverStatus struct {
	Identity       model.Identity `json:"identity"`
	Pump           string         `json:"pump"`
	PressureProven bool           `json:"pressure_proven"`
	FlowProven     bool           `json:"flow_proven"`
	Complete       bool           `json:"complete"`
	Missing        []string       `json:"missing"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Coordinator struct {
	mu              sync.RWMutex
	flow            *bearing.FlowObserver
	status          HandoverStatus
	minimumPressure float64
}

func NewCoordinator(flow *bearing.FlowObserver, minimumPressure float64) *Coordinator {
	return &Coordinator{flow: flow, minimumPressure: minimumPressure}
}

func (c *Coordinator) Begin(identity model.Identity, pump string) (HandoverStatus, error) {
	if !identity.Valid() || pump == "" {
		return HandoverStatus{}, fmt.Errorf("pump handover identity and pump are required")
	}
	c.flow.Begin(identity)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = HandoverStatus{Identity: identity, Pump: pump, Missing: []string{"supply pressure", "return flow"}, UpdatedAt: time.Now().UTC()}
	return c.status, nil
}

func (c *Coordinator) ConfirmPressure(identity model.Identity, bar float64) (HandoverStatus, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.status.Identity.SameGeneration(identity) {
		return c.status, fmt.Errorf("pressure proof belongs to another pump handover")
	}
	c.status.PressureProven = bar >= c.minimumPressure
	c.evaluateLocked()
	return c.status, nil
}

func (c *Coordinator) ConfirmReturnFlow(identity model.Identity, litresMin float64) (HandoverStatus, error) {
	flow, err := c.flow.Observe(identity, litresMin)
	if err != nil {
		return HandoverStatus{}, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status.FlowProven = flow.Proven
	c.evaluateLocked()
	return c.status, nil
}

func (c *Coordinator) evaluateLocked() {
	c.status.Complete = c.status.PressureProven
	c.status.Missing = c.status.Missing[:0]
	if !c.status.PressureProven {
		c.status.Missing = append(c.status.Missing, "supply pressure")
	}
	if !c.status.FlowProven {
		c.status.Missing = append(c.status.Missing, "return flow")
	}
	c.status.UpdatedAt = time.Now().UTC()
}

func (c *Coordinator) Snapshot() HandoverStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	status := c.status
	status.Missing = append([]string(nil), status.Missing...)
	return status
}
