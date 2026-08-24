package bypass

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type State struct {
	Identity      model.Identity `json:"identity"`
	BleedPercent  float64        `json:"bleed_percent"`
	DustFeedOpen  bool           `json:"dust_feed_open"`
	GateOpen      bool           `json:"gate_open"`
	ChamberPascal float64        `json:"chamber_pascal"`
	Status        string         `json:"status"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type Service struct {
	mu    sync.RWMutex
	state State
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Start(identity model.Identity, bleed float64) (State, error) {
	if !identity.Valid() || bleed < 0 || bleed > 100 {
		return State{}, fmt.Errorf("bypass start target is invalid")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = State{
		Identity: identity, BleedPercent: bleed, DustFeedOpen: true,
		GateOpen: true, Status: "venting", UpdatedAt: time.Now().UTC(),
	}
	return s.state, nil
}

func (s *Service) StopDustFeed(identity model.Identity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.state.Identity.SameGeneration(identity) {
		return fmt.Errorf("bypass feed stop belongs to another operation")
	}
	s.state.DustFeedOpen = false
	s.state.Status = "draining"
	s.state.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *Service) Complete(identity model.Identity) State {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Identity.SameGeneration(identity) {
		s.state.BleedPercent = 0
		s.state.GateOpen = false
		s.state.Status = "stopped"
		s.state.UpdatedAt = time.Now().UTC()
	}
	return s.state
}

func (s *Service) ObservePressure(pascal float64) State {
	s.mu.Lock()
	s.state.ChamberPascal = pascal
	s.state.UpdatedAt = time.Now().UTC()
	state := s.state
	s.mu.Unlock()
	return state
}

func (s *Service) Snapshot() State {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
