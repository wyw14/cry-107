package coalmill

import (
	"context"
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/interlock"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/operation"
)

type StartupState struct {
	Identity   model.Identity `json:"identity"`
	HotAirOpen float64        `json:"hot_air_open_percent"`
	CoalFeed   bool           `json:"coal_feed"`
	Stage      string         `json:"stage"`
	TripActive bool           `json:"trip_active"`
}

type Startup struct {
	mu          sync.RWMutex
	compensator *operation.Compensator
	trip        *interlock.COTrip
	state       StartupState
}

func NewStartup(compensator *operation.Compensator, trip *interlock.COTrip) *Startup {
	return &Startup{compensator: compensator, trip: trip}
}

func (s *Startup) Begin(identity model.Identity, previousHotAir float64) error {
	if !identity.Valid() {
		return fmt.Errorf("coal-mill startup identity is invalid")
	}
	s.mu.Lock()
	s.state = StartupState{Identity: identity, HotAirOpen: 12, Stage: "warming"}
	s.mu.Unlock()
	return s.compensator.Register("restore startup hot air", identity, func(context.Context) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.state.HotAirOpen = previousHotAir
		s.state.Stage = "rolled-back"
		return nil
	})
}

func (s *Startup) AdvanceHotAir(open float64) StartupState {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state.HotAirOpen = open
	s.state.Stage = "hot-air-ramp"
	return s.state
}

func (s *Startup) TripCO(ctx context.Context, ppm float64) (StartupState, error) {
	s.mu.Lock()
	identity := s.state.Identity
	s.mu.Unlock()
	if _, err := s.trip.CloseHotAir(identity, ppm); err != nil {
		return StartupState{}, err
	}
	s.mu.Lock()
	s.state.HotAirOpen = 0
	s.state.CoalFeed = false
	s.state.Stage = "co-trip"
	s.state.TripActive = true
	s.mu.Unlock()
	err := s.compensator.Rollback(ctx, identity)
	if err == nil {
		return s.Snapshot(), fmt.Errorf("safety terminal unexpectedly allowed startup compensation")
	}
	return s.Snapshot(), nil
}

func (s *Startup) FailOrdinary(ctx context.Context) error {
	s.mu.RLock()
	identity := s.state.Identity
	s.mu.RUnlock()
	return s.compensator.Rollback(ctx, identity)
}

func (s *Startup) Snapshot() StartupState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state
}
