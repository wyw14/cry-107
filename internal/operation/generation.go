package operation

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/model"
)

type Generations struct {
	mu      sync.RWMutex
	current model.Identity
}

func NewGenerations() *Generations {
	return &Generations{current: model.NewIdentity(1)}
}

func (g *Generations) Current() model.Identity {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.current
}

func (g *Generations) Begin() model.Identity {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.current = g.current.Next()
	return g.current
}

func (g *Generations) Accept(identity model.Identity) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if !identity.Valid() {
		return fmt.Errorf("operation identity is invalid")
	}
	if identity.RunID != g.current.RunID {
		return fmt.Errorf("operation belongs to a different kiln run")
	}
	if identity.Generation != g.current.Generation {
		return fmt.Errorf("stale operation generation %d, current %d", identity.Generation, g.current.Generation)
	}
	return nil
}

func (g *Generations) Restore(identity model.Identity) error {
	if !identity.Valid() {
		return fmt.Errorf("cannot restore invalid operation identity")
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	if identity.Generation < g.current.Generation {
		return fmt.Errorf("cannot restore older generation")
	}
	g.current = identity
	return nil
}
