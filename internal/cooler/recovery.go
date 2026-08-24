package cooler

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/crusher"
	"github.com/wyw14/cry-107/internal/model"
)

type RecoveryState struct {
	Identity  model.Identity          `json:"identity"`
	Stage     string                  `json:"stage"`
	Transport conveyor.TransportState `json:"transport"`
	Crusher   crusher.PermitState     `json:"crusher"`
	Complete  bool                    `json:"complete"`
	UpdatedAt time.Time               `json:"updated_at"`
}

type RecoveryGraph struct {
	mu       sync.RWMutex
	conveyor *conveyor.StateService
	crusher  *crusher.Permit
	state    RecoveryState
}

func NewRecoveryGraph(transport *conveyor.StateService, permit *crusher.Permit) *RecoveryGraph {
	return &RecoveryGraph{conveyor: transport, crusher: permit}
}

func (g *RecoveryGraph) Begin(identity model.Identity, backlog float64) RecoveryState {
	transport := g.conveyor.BeginRecovery(identity, backlog)
	g.mu.Lock()
	g.state = RecoveryState{Identity: identity, Stage: "conveyor-start", Transport: transport, UpdatedAt: time.Now().UTC()}
	state := g.state
	g.mu.Unlock()
	return state
}

func (g *RecoveryGraph) ApplySpeedProof(identity model.Identity, speed float64) (RecoveryState, error) {
	transport, err := g.conveyor.MarkRunning(identity, speed)
	if err != nil {
		return RecoveryState{}, err
	}
	return g.evaluate(identity, transport), nil
}

func (g *RecoveryGraph) ReleaseBacklog(identity model.Identity, tonnes float64) (RecoveryState, error) {
	transport, err := g.conveyor.ReleaseBacklog(identity, tonnes)
	if err != nil {
		return RecoveryState{}, err
	}
	return g.evaluate(identity, transport), nil
}

func (g *RecoveryGraph) evaluate(identity model.Identity, transport conveyor.TransportState) RecoveryState {
	permit := g.crusher.Evaluate(identity, transport)
	g.mu.Lock()
	defer g.mu.Unlock()
	g.state.Transport = transport
	g.state.Crusher = permit
	g.state.UpdatedAt = time.Now().UTC()
	if permit.Allowed {
		g.state.Stage = "crusher-start"
		g.state.Complete = true
	} else if transport.Running {
		g.state.Stage = "controlled-backlog-release"
	}
	return g.state
}

func (g *RecoveryGraph) RequireComplete() error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !g.state.Complete {
		return fmt.Errorf("cooler recovery incomplete: %s", g.state.Crusher.Reason)
	}
	return nil
}

func (g *RecoveryGraph) Snapshot() RecoveryState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}
