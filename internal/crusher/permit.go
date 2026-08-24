package crusher

import (
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/conveyor"
	"github.com/wyw14/cry-107/internal/model"
)

type PermitState struct {
	Identity      model.Identity `json:"identity"`
	Allowed       bool           `json:"allowed"`
	Reason        string         `json:"reason"`
	ConveyorReady bool           `json:"conveyor_ready"`
	BacklogClear  bool           `json:"backlog_clear"`
	EvaluatedAt   time.Time      `json:"evaluated_at"`
}

type Permit struct {
	mu    sync.RWMutex
	state PermitState
}

func NewPermit() *Permit {
	return &Permit{}
}

func (p *Permit) Evaluate(identity model.Identity, transport conveyor.TransportState) PermitState {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = PermitState{
		Identity: identity, ConveyorReady: transport.Running,
		BacklogClear: transport.BacklogDrained, EvaluatedAt: time.Now().UTC(),
	}
	switch {
	case !transport.Identity.SameGeneration(identity):
		p.state.Reason = "conveyor proof belongs to another recovery"
	case !transport.Running:
		p.state.Reason = "conveyor speed not established"
	case !transport.BacklogDrained:
		p.state.Reason = "clinker backlog still releasing"
	default:
		p.state.Allowed = true
		p.state.Reason = "transport and material paths ready"
	}
	return p.state
}

func (p *Permit) Snapshot() PermitState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}
