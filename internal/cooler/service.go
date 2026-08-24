package cooler

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/grate"
	"github.com/wyw14/cry-107/internal/model"
)

type State struct {
	Identity      model.Identity `json:"identity"`
	InletTemp     float64        `json:"inlet_temp_c"`
	OutletTemp    float64        `json:"outlet_temp_c"`
	CoolingAir    float64        `json:"cooling_air"`
	SafetyMinimum float64        `json:"safety_minimum"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type Service struct {
	mu        sync.RWMutex
	allocator *Allocator
	fan       *grate.Fan
	state     State
}

func NewService(allocator *Allocator, fan *grate.Fan) *Service {
	return &Service{allocator: allocator, fan: fan}
}

func (s *Service) Adjust(identity model.Identity, outletTemp, requestedAir float64) (State, error) {
	minimum := 45.0
	if outletTemp > 135 {
		minimum = 65
	}
	allocation, err := s.allocator.Request(identity, requestedAir, minimum)
	if err != nil {
		return State{}, fmt.Errorf("allocate clinker cooling air: %w", err)
	}
	if _, err := s.fan.Command(identity, allocation.Granted); err != nil {
		return State{}, err
	}
	state := State{
		Identity: identity, InletTemp: 1320, OutletTemp: outletTemp,
		CoolingAir: allocation.Granted, SafetyMinimum: minimum, UpdatedAt: time.Now().UTC(),
	}
	s.mu.Lock()
	s.state = state
	s.mu.Unlock()
	return state, nil
}

func (s *Service) Snapshot() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
