package coalmill

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/gas"
)

type InertingPermit interface {
	InertingStarted(uuid.UUID)
	InertingCompleted(uuid.UUID) bool
}

type InertingState struct {
	SessionID      uuid.UUID       `json:"session_id"`
	NitrogenVolume float64         `json:"nitrogen_volume_m3"`
	RequiredVolume float64         `json:"required_volume_m3"`
	Oxygen         gas.OxygenState `json:"oxygen"`
	Complete       bool            `json:"complete"`
	StartedAt      time.Time       `json:"started_at"`
	CompletedAt    time.Time       `json:"completed_at,omitempty"`
}

type InertingService struct {
	mu     sync.RWMutex
	window *gas.OxygenWindow
	permit InertingPermit
	state  InertingState
}

func NewInertingService(window *gas.OxygenWindow, permit InertingPermit, requiredVolume float64) *InertingService {
	return &InertingService{window: window, permit: permit, state: InertingState{RequiredVolume: requiredVolume}}
}

func (s *InertingService) Begin() InertingState {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := uuid.New()
	s.window.Reset(session)
	s.permit.InertingStarted(session)
	s.state = InertingState{SessionID: session, RequiredVolume: s.state.RequiredVolume, StartedAt: time.Now().UTC()}
	return s.state
}

func (s *InertingService) AddNitrogen(session uuid.UUID, cubicMetres float64) (InertingState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session != s.state.SessionID || cubicMetres < 0 {
		return s.state, fmt.Errorf("nitrogen flow belongs to another inerting session")
	}
	s.state.NitrogenVolume += cubicMetres
	s.evaluateLocked()
	return s.state, nil
}

func (s *InertingService) ObserveOxygen(sample gas.OxygenSample) (InertingState, error) {
	state, err := s.window.Add(sample)
	if err != nil {
		return InertingState{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Oxygen = state
	s.evaluateLocked()
	return s.state, nil
}

func (s *InertingService) evaluateLocked() {
	if s.state.Complete || s.state.NitrogenVolume < s.state.RequiredVolume || !s.state.Oxygen.Stable {
		return
	}
	if s.permit.InertingCompleted(s.state.SessionID) {
		s.state.Complete = true
		s.state.CompletedAt = time.Now().UTC()
	}
}

func (s *InertingService) Snapshot() InertingState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
