package wasteheat

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/grate"
	"github.com/wyw14/cry-107/internal/model"
)

type Demand struct {
	Identity       model.Identity `json:"identity"`
	Requested      float64        `json:"requested"`
	Granted        float64        `json:"granted"`
	BudgetRevision uint64         `json:"budget_revision"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type DemandService struct {
	mu      sync.RWMutex
	budget  *grate.FanBudget
	current Demand
}

func NewDemandService(budget *grate.FanBudget) *DemandService {
	return &DemandService{budget: budget}
}

func (s *DemandService) Request(identity model.Identity, requested float64) (Demand, error) {
	if !identity.Valid() {
		return Demand{}, fmt.Errorf("waste-heat demand identity is invalid")
	}
	budget, err := s.budget.ReserveWasteHeat(requested)
	if err != nil {
		return Demand{}, err
	}
	demand := Demand{
		Identity: identity, Requested: requested, Granted: budget.WasteHeat,
		BudgetRevision: budget.Revision, UpdatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.current = demand
	s.mu.Unlock()
	return demand, nil
}

func (s *DemandService) Snapshot() Demand {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current
}
