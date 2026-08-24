package bypass

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/sampler"
)

type EnrichmentResult struct {
	BatchID    uuid.UUID `json:"batch_id"`
	Value      float64   `json:"value"`
	GasPPM     float64   `json:"gas_ppm"`
	DustMassKG float64   `json:"dust_mass_kg"`
	Calculated time.Time `json:"calculated_at"`
}

type Enrichment struct {
	mu         sync.RWMutex
	aggregator *sampler.Aggregator
	latest     EnrichmentResult
}

func NewEnrichment(aggregator *sampler.Aggregator) *Enrichment {
	return &Enrichment{aggregator: aggregator}
}

func (e *Enrichment) Calculate(batchID uuid.UUID) (EnrichmentResult, error) {
	batch, complete := e.aggregator.Complete(batchID)
	if !complete {
		return EnrichmentResult{}, fmt.Errorf("residence batch %s is awaiting paired gas and dust samples", batchID)
	}
	value := (batch.Gas.Chlorine * batch.Dust.Chloride) / batch.Dust.MassKG
	result := EnrichmentResult{
		BatchID: batchID, Value: value, GasPPM: batch.Gas.Chlorine,
		DustMassKG: batch.Dust.MassKG, Calculated: time.Now().UTC(),
	}
	e.mu.Lock()
	e.latest = result
	e.mu.Unlock()
	return result, nil
}

func (e *Enrichment) Latest() EnrichmentResult {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.latest
}
