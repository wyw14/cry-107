package feeder

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/model"
)

type Permit struct {
	mu       sync.RWMutex
	identity model.Identity
	enabled  bool
	reason   string
	revision uint64
}

func NewPermit() *Permit {
	return &Permit{}
}

func (p *Permit) Grant(identity model.Identity, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.identity = identity
	p.enabled = true
	p.reason = reason
	p.revision++
}

func (p *Permit) Revoke(identity model.Identity, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.identity.SameGeneration(identity) {
		p.enabled = false
		p.reason = reason
		p.revision++
	}
}

func (p *Permit) Check(identity model.Identity) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.enabled {
		return fmt.Errorf("feed is interlocked: %s", p.reason)
	}
	if !p.identity.SameGeneration(identity) {
		return fmt.Errorf("feed permit generation mismatch")
	}
	return nil
}

func (p *Permit) State() (bool, string, uint64) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.enabled, p.reason, p.revision
}
