package cooler

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/grate"
	"github.com/wyw14/cry-107/internal/model"
)

type Allocation struct {
	Identity       model.Identity `json:"identity"`
	Requested      float64        `json:"requested"`
	Minimum        float64        `json:"minimum"`
	Granted        float64        `json:"granted"`
	BudgetRevision uint64         `json:"budget_revision"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type Allocator struct {
	mu      sync.RWMutex
	budget  *grate.FanBudget
	current Allocation
}

func NewAllocator(budget *grate.FanBudget) *Allocator {
	return &Allocator{budget: budget}
}

func (a *Allocator) Request(identity model.Identity, requested, minimum float64) (Allocation, error) {
	if !identity.Valid() {
		return Allocation{}, fmt.Errorf("cooling allocation identity is invalid")
	}
	budget, err := a.budget.ReserveCooling(requested, minimum)
	if err != nil {
		return Allocation{}, err
	}
	allocation := Allocation{
		Identity: identity, Requested: requested, Minimum: minimum,
		Granted: budget.Cooling, BudgetRevision: budget.Revision, UpdatedAt: time.Now().UTC(),
	}
	a.mu.Lock()
	a.current = allocation
	a.mu.Unlock()
	return allocation, nil
}

func (a *Allocator) Snapshot() Allocation {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.current
}
