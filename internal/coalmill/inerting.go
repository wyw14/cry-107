package coalmill

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/gas"
)

// InertingPermit is the bridge between the inerting state machine and the
// coal-mill interlock that authorises hot air and coal feed. The state
// machine calls InertingStarted when a purge begins, InertingCompleted once
// the purge volume and the continuous low-oxygen window are both proven, and
// InertingRevoked the instant either condition no longer holds. Completion is
// not a latching relay: the permit must track the live oxygen window, so the
// hot-air damper closes as soon as oxygen creeps back above the limit instead
// of waiting for a downstream high-oxygen interlock to trip.
type InertingPermit interface {
	InertingStarted(uuid.UUID)
	InertingCompleted(uuid.UUID) bool
	InertingRevoked(uuid.UUID, string)
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

// evaluateLocked recomputes the completion state from the live process
// conditions on every sample. Completion is not a latching relay: it requires
// both the required purge volume AND a continuous low-oxygen window (the window
// reports Stable only once oxygen has stayed at or below the limit for the
// configured duration). A single instantaneous low reading can never satisfy
// the window, and any frame that pushes oxygen back above the limit breaks the
// window immediately. When that happens after inerting had completed, the
// permit is revoked so the hot-air damper and coal feed are closed before a
// downstream high-oxygen interlock has to fire.
func (s *InertingService) evaluateLocked() {
	volumeProven := s.state.NitrogenVolume >= s.state.RequiredVolume
	windowProven := s.state.Oxygen.Stable && s.state.Oxygen.StableFor > 0
	proven := volumeProven && windowProven
	switch {
	case proven && !s.state.Complete:
		if s.permit.InertingCompleted(s.state.SessionID) {
			s.state.Complete = true
			s.state.CompletedAt = time.Now().UTC()
		}
	case !proven && s.state.Complete:
		s.state.Complete = false
		s.state.CompletedAt = time.Time{}
		s.permit.InertingRevoked(s.state.SessionID, "inerting window broken")
	}
}

func (s *InertingService) Snapshot() InertingState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
