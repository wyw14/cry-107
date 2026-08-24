package feeder

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/rawmix"
)

type State struct {
	Identity       model.Identity `json:"identity"`
	TargetTPH      float64        `json:"target_tph"`
	ActualTPH      float64        `json:"actual_tph"`
	Running        bool           `json:"running"`
	LastAdjustment time.Time      `json:"last_adjustment"`
}

type Service struct {
	mu     sync.RWMutex
	state  State
	permit *Permit
}

func NewService(permit *Permit) *Service {
	return &Service{permit: permit}
}

func (s *Service) Apply(target rawmix.Target) (State, error) {
	if err := s.permit.Check(target.Identity); err != nil {
		return State{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Identity.Generation > target.Identity.Generation {
		return State{}, fmt.Errorf("feed target belongs to stale operation generation")
	}
	s.state.Identity = target.Identity
	s.state.TargetTPH = target.TonnesPerHour
	s.state.Running = target.TonnesPerHour > 0
	s.state.LastAdjustment = time.Now().UTC()
	return s.state, nil
}

func (s *Service) Observe(actualTPH float64, identity model.Identity) (State, error) {
	if actualTPH < 0 {
		return State{}, fmt.Errorf("feeder observation cannot be negative")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.state.Identity.SameGeneration(identity) {
		return State{}, fmt.Errorf("feeder observation belongs to another operation")
	}
	s.state.ActualTPH = actualTPH
	s.state.LastAdjustment = time.Now().UTC()
	return s.state, nil
}

func (s *Service) Stop(identity model.Identity) State {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Identity.SameGeneration(identity) {
		s.state.TargetTPH = 0
		s.state.Running = false
		s.state.LastAdjustment = time.Now().UTC()
	}
	return s.state
}

func (s *Service) Snapshot() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
