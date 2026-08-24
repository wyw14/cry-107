package pressure

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type GateState struct {
	Identity model.Identity `json:"identity"`
	Open     bool           `json:"open"`
	Proof    bool           `json:"proof"`
	Changed  time.Time      `json:"changed_at"`
}

type Gate struct {
	mu    sync.RWMutex
	state GateState
}

func NewGate(open bool) *Gate {
	return &Gate{state: GateState{Open: open, Proof: true, Changed: time.Now().UTC()}}
}

func (g *Gate) Set(identity model.Identity, open bool) (GateState, error) {
	if !identity.Valid() {
		return GateState{}, fmt.Errorf("gate command identity is invalid")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.state = GateState{Identity: identity, Open: open, Proof: false, Changed: time.Now().UTC()}
	return g.state, nil
}

func (g *Gate) Confirm(identity model.Identity, open bool) (GateState, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.state.Identity.SameGeneration(identity) || g.state.Open != open {
		return g.state, fmt.Errorf("gate proof does not match command")
	}
	g.state.Proof = true
	g.state.Changed = time.Now().UTC()
	return g.state, nil
}

func (g *Gate) Snapshot() GateState {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.state
}
