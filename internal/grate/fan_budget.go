package grate

import (
	"fmt"
	"sync"
	"time"
)

type BudgetState struct {
	Capacity       float64   `json:"capacity"`
	Cooling        float64   `json:"cooling"`
	WasteHeat      float64   `json:"waste_heat"`
	CoolingMinimum float64   `json:"cooling_minimum"`
	Revision       uint64    `json:"revision"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (s BudgetState) Total() float64 {
	return s.Cooling + s.WasteHeat
}

type FanBudget struct {
	mu    sync.Mutex
	state BudgetState
}

func NewFanBudget(capacity float64) *FanBudget {
	return &FanBudget{state: BudgetState{Capacity: capacity, UpdatedAt: time.Now().UTC()}}
}

func (b *FanBudget) ReserveCooling(request, minimum float64) (BudgetState, error) {
	if request < 0 || minimum < 0 || minimum > request {
		return BudgetState{}, fmt.Errorf("cooling fan request is invalid")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if minimum > b.state.Capacity {
		return b.state, fmt.Errorf("cooling safety minimum exceeds fan capacity")
	}
	b.state.CoolingMinimum = minimum
	b.state.Cooling = min(request, b.state.Capacity)
	remaining := b.state.Capacity - b.state.Cooling
	if b.state.WasteHeat > remaining {
		b.state.WasteHeat = remaining
	}
	b.commitLocked()
	return b.state, nil
}

func (b *FanBudget) ReserveWasteHeat(request float64) (BudgetState, error) {
	if request < 0 {
		return BudgetState{}, fmt.Errorf("waste-heat fan request is invalid")
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	reservedCooling := b.state.Cooling
	if reservedCooling < b.state.CoolingMinimum {
		reservedCooling = b.state.CoolingMinimum
	}
	available := b.state.Capacity - reservedCooling
	if available < 0 {
		available = 0
	}
	b.state.WasteHeat = min(request, available)
	b.commitLocked()
	return b.state, nil
}

func (b *FanBudget) commitLocked() {
	b.state.Revision++
	b.state.UpdatedAt = time.Now().UTC()
}

func (b *FanBudget) Snapshot() BudgetState {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}
