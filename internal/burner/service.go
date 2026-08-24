package burner

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/model"
)

type Service struct {
	mu        sync.Mutex
	loop      *RatioLoop
	commander *Commander
}

func NewService(loop *RatioLoop, commander *Commander) *Service {
	return &Service{loop: loop, commander: commander}
}

func (s *Service) RequestLoad(identity model.Identity, load float64) (RatioPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.loop.Propose(identity, load)
}

func (s *Service) AcceptAir(identity model.Identity, target float64) (RatioState, error) {
	if err := s.loop.AirCommandAccepted(identity, target); err != nil {
		return RatioState{}, err
	}
	return s.loop.Snapshot(), nil
}

func (s *Service) ConfirmAir(identity model.Identity, actual float64) (Command, error) {
	if err := s.loop.AirConfirmed(identity, actual); err != nil {
		return Command{}, err
	}
	state := s.loop.Snapshot()
	command := Command{
		Identity: identity, FuelTPH: state.FuelTarget,
		PrimaryAir: actual * 0.12, SecondaryAir: actual,
	}
	if err := s.commander.Issue(command); err != nil {
		return Command{}, fmt.Errorf("issue confirmed burner target: %w", err)
	}
	return s.commander.Current(), nil
}

func (s *Service) State() map[string]any {
	return map[string]any{"ratio": s.loop.Snapshot(), "command": s.commander.Current()}
}

func (s *Service) Close(identity model.Identity) Command {
	return s.commander.Close(identity)
}
