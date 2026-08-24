package conveyor

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type TransportState struct {
	Identity       model.Identity `json:"identity"`
	SpeedPercent   float64        `json:"speed_percent"`
	Running        bool           `json:"running"`
	BacklogTonnes  float64        `json:"backlog_tonnes"`
	BacklogDrained bool           `json:"backlog_drained"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type StateService struct {
	mu    sync.RWMutex
	state TransportState
}

func NewStateService() *StateService {
	return &StateService{}
}

func (s *StateService) BeginRecovery(identity model.Identity, backlog float64) TransportState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = TransportState{Identity: identity, BacklogTonnes: backlog, BacklogDrained: backlog == 0, UpdatedAt: time.Now().UTC()}
	return s.state
}

func (s *StateService) MarkRunning(identity model.Identity, speed float64) (TransportState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.state.Identity.SameGeneration(identity) {
		return s.state, fmt.Errorf("conveyor speed proof belongs to another recovery")
	}
	s.state.SpeedPercent = speed
	s.state.Running = speed >= 95
	s.state.UpdatedAt = time.Now().UTC()
	return s.state, nil
}

func (s *StateService) ReleaseBacklog(identity model.Identity, tonnes float64) (TransportState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.state.Identity.SameGeneration(identity) || tonnes < 0 {
		return s.state, fmt.Errorf("backlog release is invalid")
	}
	s.state.BacklogTonnes -= tonnes
	if s.state.BacklogTonnes <= 0 {
		s.state.BacklogTonnes = 0
		s.state.BacklogDrained = true
	}
	s.state.UpdatedAt = time.Now().UTC()
	return s.state, nil
}

func (s *StateService) Snapshot() TransportState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
