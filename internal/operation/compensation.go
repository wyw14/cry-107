package operation

import (
	"context"
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/model"
)

type Compensation func(context.Context) error

type compensationEntry struct {
	name       string
	identity   model.Identity
	compensate Compensation
}

type Compensator struct {
	mu       sync.Mutex
	entries  []compensationEntry
	barriers map[uint64]string
}

func NewCompensator() *Compensator {
	return &Compensator{barriers: make(map[uint64]string)}
}

func (c *Compensator) Register(name string, identity model.Identity, fn Compensation) error {
	if name == "" || fn == nil || !identity.Valid() {
		return fmt.Errorf("compensation registration is incomplete")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if reason, blocked := c.barriers[identity.Generation]; blocked {
		return fmt.Errorf("operation is behind safety barrier: %s", reason)
	}
	c.entries = append(c.entries, compensationEntry{name: name, identity: identity, compensate: fn})
	return nil
}

func (c *Compensator) EstablishBarrier(generation uint64, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.barriers[generation] = reason
}

func (c *Compensator) Rollback(ctx context.Context, identity model.Identity) error {
	c.mu.Lock()
	entries := append([]compensationEntry(nil), c.entries...)
	barrier := c.barriers[identity.Generation]
	c.mu.Unlock()
	if barrier != "" {
		return fmt.Errorf("rollback blocked by safety terminal: %s", barrier)
	}
	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		if !entry.identity.SameGeneration(identity) {
			continue
		}
		if err := entry.compensate(ctx); err != nil {
			return fmt.Errorf("compensation %s failed: %w", entry.name, err)
		}
	}
	return nil
}

func (c *Compensator) Pending(identity model.Identity) []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	var names []string
	for _, entry := range c.entries {
		if entry.identity.SameGeneration(identity) {
			names = append(names, entry.name)
		}
	}
	return names
}
