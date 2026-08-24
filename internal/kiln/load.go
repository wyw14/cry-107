package kiln

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/burner"
	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/pressure"
)

type LoadCoordinator struct {
	mu      sync.Mutex
	runtime *Runtime
	damper  *pressure.Damper
	burner  *burner.Service
}

func NewLoadCoordinator(runtime *Runtime, damper *pressure.Damper, burnerService *burner.Service) *LoadCoordinator {
	return &LoadCoordinator{runtime: runtime, damper: damper, burner: burnerService}
}

func (c *LoadCoordinator) Request(identity model.Identity, target float64) (burner.RatioPlan, pressure.DamperState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	plan, err := c.burner.RequestLoad(identity, target)
	if err != nil {
		return burner.RatioPlan{}, pressure.DamperState{}, err
	}
	damper, err := c.damper.Command(identity, plan.RequiredAir)
	if err != nil {
		return burner.RatioPlan{}, pressure.DamperState{}, err
	}
	if _, err := c.burner.AcceptAir(identity, plan.RequiredAir); err != nil {
		return burner.RatioPlan{}, pressure.DamperState{}, err
	}
	return plan, damper, nil
}

func (c *LoadCoordinator) ConfirmAir(identity model.Identity, actual float64) (RuntimeState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, err := c.damper.Confirm(identity, actual); err != nil {
		return RuntimeState{}, fmt.Errorf("confirm combustion damper: %w", err)
	}
	command, err := c.burner.ConfirmAir(identity, actual)
	if err != nil {
		return RuntimeState{}, fmt.Errorf("advance burner fuel: %w", err)
	}
	current := c.runtime.Snapshot()
	return c.runtime.ApplyTargets(current.SpeedRPM, current.FeedTPH, command.FuelTPH/0.22, identity)
}

func (c *LoadCoordinator) State() map[string]any {
	return map[string]any{
		"kiln":   c.runtime.Snapshot(),
		"damper": c.damper.Snapshot(),
		"burner": c.burner.State(),
	}
}
