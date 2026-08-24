package model

import (
	"errors"
	"fmt"
	"time"
)

type KilnPhase string

const (
	PhasePreheat  KilnPhase = "preheat"
	PhaseFeeding  KilnPhase = "feeding"
	PhaseBurning  KilnPhase = "burning"
	PhaseCooling  KilnPhase = "cooling"
	PhaseShutdown KilnPhase = "shutdown"
	PhaseRecovery KilnPhase = "recovery"
)

var phaseOrder = map[KilnPhase]int{
	PhasePreheat:  0,
	PhaseFeeding:  1,
	PhaseBurning:  2,
	PhaseCooling:  3,
	PhaseShutdown: 4,
	PhaseRecovery: 5,
}

type ProcessState struct {
	Identity  Identity  `json:"identity"`
	Phase     KilnPhase `json:"phase"`
	UpdatedAt time.Time `json:"updated_at"`
	Reason    string    `json:"reason"`
}

func NewProcessState() ProcessState {
	return ProcessState{Identity: NewIdentity(1), Phase: PhasePreheat, UpdatedAt: time.Now().UTC()}
}

func (s ProcessState) Transition(next KilnPhase, reason string) (ProcessState, error) {
	currentIndex, currentOK := phaseOrder[s.Phase]
	nextIndex, nextOK := phaseOrder[next]
	if !currentOK || !nextOK {
		return s, fmt.Errorf("unknown kiln phase transition %q to %q", s.Phase, next)
	}
	allowed := nextIndex == currentIndex+1 || (s.Phase == PhaseRecovery && next == PhasePreheat)
	if next == PhaseShutdown && s.Phase != PhaseShutdown {
		allowed = true
	}
	if !allowed {
		return s, errors.New("kiln phase transition violates lifecycle order")
	}
	s.Phase = next
	s.Reason = reason
	s.UpdatedAt = time.Now().UTC()
	s.Identity = s.Identity.Next()
	return s, nil
}

func (s ProcessState) Terminal() bool {
	return s.Phase == PhaseShutdown
}
