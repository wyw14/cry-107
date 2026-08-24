package pressure

import (
	"fmt"
	"sync"
	"time"
)

type Mode string

const (
	ModeManual Mode = "manual"
	ModeAuto   Mode = "auto"
)

type ModeState struct {
	Mode       Mode      `json:"mode"`
	Setpoint   float64   `json:"setpoint"`
	Damper     float64   `json:"damper"`
	ChangedAt  time.Time `json:"changed_at"`
	ChangeNote string    `json:"change_note"`
}

type ModeService struct {
	mu         sync.RWMutex
	controller *Controller
	state      ModeState
}

func NewModeService(controller *Controller, damper float64) *ModeService {
	return &ModeService{controller: controller, state: ModeState{Mode: ModeAuto, Damper: damper, ChangedAt: time.Now().UTC()}}
}

func (s *ModeService) EnterManual(currentDamper, processValue float64) ModeState {
	s.controller.TrackManual(currentDamper, processValue)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Mode = ModeManual
	s.state.Damper = currentDamper
	s.state.ChangeNote = "operator control"
	s.state.ChangedAt = time.Now().UTC()
	return s.state
}

func (s *ModeService) ManualMove(target, processValue float64) (ModeState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Mode != ModeManual {
		return s.state, fmt.Errorf("manual movement requires manual mode")
	}
	s.state.Damper = target
	s.state.ChangedAt = time.Now().UTC()
	s.controller.TrackManual(target, processValue)
	return s.state, nil
}

func (s *ModeService) EnterAuto(setpoint, processValue, currentDamper float64) ModeState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.Mode = ModeAuto
	s.state.Setpoint = setpoint
	s.state.Damper = currentDamper
	s.state.ChangeNote = "bumpless automatic transfer"
	s.state.ChangedAt = time.Now().UTC()
	return s.state
}

func (s *ModeService) Tick(processValue float64) (ModeState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state.Mode != ModeAuto {
		return s.state, fmt.Errorf("automatic tick requires automatic mode")
	}
	controller := s.controller.Tick(processValue)
	s.state.Damper = controller.Output
	return s.state, nil
}

func (s *ModeService) Snapshot() ModeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
