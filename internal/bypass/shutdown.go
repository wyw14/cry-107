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

// drainTimeout bounds how long the dust conveyor may take to empty once the
// feed is stopped. It is deliberately generous relative to the configured
// drain rate, but it must remain a real deadline: if the path cannot empty in
// time the shutdown fails rather than closing the gate on a charged conveyor.
const drainTimeout = 5 * time.Second

// Run executes the bypass shutdown in the order the physical path demands:
// stop new dust first, then empty the conveyor, and only close the discharge
// gate once nothing remains to back-flow into the kiln tail. Closing the gate
// while the conveyor is still charged is what traps material and causes the
// chamber pressure to spike; draining first removes the source. A drain that
// times out is a genuine failure — the gate stays open and the operation is
// not reported complete.
func (s *Shutdown) Run(ctx context.Context, identity model.Identity) (ShutdownResult, error) {
	result := ShutdownResult{Identity: identity}
	if err := s.service.StopDustFeed(identity); err != nil {
		return result, fmt.Errorf("stop new bypass dust: %w", err)
	}
	result.DustFeedStopped = true
	drainContext, cancel := context.WithTimeout(ctx, drainTimeout)
	defer cancel()
	drain, err := s.drainer.Drain(drainContext, identity)
	result.Drain = drain
	if err != nil {
		// The conveyor could not empty in time. Keep the discharge gate open
		// so trapped dust is not sealed into the path, and surface this as a
		// real failure rather than marking an un-drained shutdown complete.
		return result, fmt.Errorf("bypass conveyor drain failed before gate close: %w", err)
	}
	gate, err := s.gate.Set(identity, false)
	if err != nil {
		return result, fmt.Errorf("command bypass discharge gate: %w", err)
	}
	gate, err = s.gate.Confirm(identity, false)
	if err != nil {
		return result, fmt.Errorf("confirm bypass discharge gate: %w", err)
	}
	result.Gate = gate
	result.Complete = true
	result.CompletedAt = time.Now().UTC()
	s.service.Complete(identity)
	return result, nil
}
