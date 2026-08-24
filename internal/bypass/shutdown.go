package bypass

import (
	"context"
	"fmt"
	"time"

	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/pressure"
)

type ShutdownResult struct {
	Identity        model.Identity      `json:"identity"`
	DustFeedStopped bool                `json:"dust_feed_stopped"`
	Drain           conveyor.DrainState `json:"drain"`
	Gate            pressure.GateState  `json:"gate"`
	Complete        bool                `json:"complete"`
	CompletedAt     time.Time           `json:"completed_at,omitempty"`
}

type Shutdown struct {
	service *Service
	drainer *conveyor.Drainer
	gate    *pressure.Gate
}

func NewShutdown(service *Service, drainer *conveyor.Drainer, gate *pressure.Gate) *Shutdown {
	return &Shutdown{service: service, drainer: drainer, gate: gate}
}

func (s *Shutdown) Prepare(identity model.Identity, kilograms float64) error {
	if !identity.Valid() || kilograms < 0 {
		return fmt.Errorf("bypass drain preparation is invalid")
	}
	s.drainer.Load(identity, kilograms)
	return nil
}

func (s *Shutdown) Run(ctx context.Context, identity model.Identity) (ShutdownResult, error) {
	result := ShutdownResult{Identity: identity}
	if err := s.service.StopDustFeed(identity); err != nil {
		return result, fmt.Errorf("stop new bypass dust: %w", err)
	}
	result.DustFeedStopped = true
	gate, err := s.gate.Set(identity, false)
	if err != nil {
		return result, fmt.Errorf("command bypass discharge gate: %w", err)
	}
	gate, err = s.gate.Confirm(identity, false)
	if err != nil {
		return result, fmt.Errorf("confirm bypass discharge gate: %w", err)
	}
	result.Gate = gate
	drainContext, cancel := context.WithCancel(ctx)
	cancel()
	drain, err := s.drainer.Drain(drainContext, identity)
	result.Drain = drain
	if err != nil {
		return result, fmt.Errorf("bypass drain failed after gate close: %w", err)
	}
	result.Complete = true
	result.CompletedAt = time.Now().UTC()
	s.service.Complete(identity)
	return result, nil
}
