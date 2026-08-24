package kiln

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type CirculationDecision struct {
	BatchID    uuid.UUID `json:"batch_id"`
	Enrichment float64   `json:"enrichment"`
	BleedOpen  float64   `json:"bleed_open_percent"`
	AppliedAt  time.Time `json:"applied_at"`
}

type Circulation struct {
	mu       sync.RWMutex
	decision CirculationDecision
}

func NewCirculation() *Circulation {
	return &Circulation{}
}

func (c *Circulation) ApplyBleed(batch uuid.UUID, enrichment float64) (CirculationDecision, error) {
	if batch == uuid.Nil || enrichment < 0 {
		return CirculationDecision{}, fmt.Errorf("chlorine circulation decision is invalid")
	}
	bleed := 20 + enrichment*18
	if bleed > 90 {
		bleed = 90
	}
	decision := CirculationDecision{
		BatchID: batch, Enrichment: enrichment, BleedOpen: bleed, AppliedAt: time.Now().UTC(),
	}
	c.mu.Lock()
	c.decision = decision
	c.mu.Unlock()
	return decision, nil
}

func (c *Circulation) Snapshot() CirculationDecision {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.decision
}
