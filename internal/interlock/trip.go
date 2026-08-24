package interlock

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-107/internal/journal"
	"github.com/wyw14/cry-107/internal/model"
)

type TripActions interface {
	StopFeed(context.Context, model.Identity) error
	CloseFuel(context.Context, model.Identity) error
	HoldCooling(context.Context, model.Identity) error
}

type TripService struct {
	manager *Manager
	journal *journal.Store
	actions TripActions
}

func NewTripService(manager *Manager, store *journal.Store, actions TripActions) *TripService {
	return &TripService{manager: manager, journal: store, actions: actions}
}

func (s *TripService) Execute(ctx context.Context, identity model.Identity, source, reason string) (Incident, error) {
	incident, err := s.manager.Raise(identity, source, "trip", reason, true)
	if err != nil {
		return Incident{}, err
	}
	steps := []struct {
		name string
		run  func(context.Context, model.Identity) error
	}{
		{name: "stop feed", run: s.actions.StopFeed},
		{name: "close fuel", run: s.actions.CloseFuel},
		{name: "hold cooling", run: s.actions.HoldCooling},
	}
	for _, step := range steps {
		if err := step.run(ctx, identity); err != nil {
			return incident, fmt.Errorf("trip action %s: %w", step.name, err)
		}
	}
	event, err := model.NewEvent("interlock.trip", identity, incident)
	if err != nil {
		return incident, err
	}
	if err := s.journal.Append(event); err != nil {
		return incident, fmt.Errorf("persist trip terminal: %w", err)
	}
	return incident, nil
}
