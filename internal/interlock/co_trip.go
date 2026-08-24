package interlock

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-107/internal/model"
	"github.com/wyw14/cry-107/internal/operation"
)

type COTripState struct {
	Identity     model.Identity `json:"identity"`
	Active       bool           `json:"active"`
	HotAirOpen   float64        `json:"hot_air_open_percent"`
	CoalFeedOpen bool           `json:"coal_feed_open"`
	Reason       string         `json:"reason"`
	TrippedAt    time.Time      `json:"tripped_at"`
}

type COTrip struct {
	manager     *Manager
	permit      *MillPermit
	compensator *operation.Compensator
	state       COTripState
}

func NewCOTrip(manager *Manager, permit *MillPermit, compensator *operation.Compensator) *COTrip {
	return &COTrip{manager: manager, permit: permit, compensator: compensator}
}

func (t *COTrip) CloseHotAir(identity model.Identity, ppm float64) (COTripState, error) {
	if ppm <= 0 {
		return COTripState{}, fmt.Errorf("CO trip needs a positive observation")
	}
	incident, err := t.manager.Raise(identity, "coal-mill-co", "trip", fmt.Sprintf("CO high-high %.0f ppm", ppm), true)
	if err != nil {
		return COTripState{}, err
	}
	t.permit.Trip(incident.Reason)
	t.state = COTripState{
		Identity: identity, Active: true, HotAirOpen: 0, CoalFeedOpen: false,
		Reason: incident.Reason, TrippedAt: time.Now().UTC(),
	}
	return t.state, nil
}

func (t *COTrip) Snapshot() COTripState {
	return t.state
}
