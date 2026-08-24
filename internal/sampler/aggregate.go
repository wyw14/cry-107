package sampler

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type GasSample struct {
	BatchID  uuid.UUID `json:"batch_id"`
	Chlorine float64   `json:"chlorine_ppm"`
	Observed time.Time `json:"observed_at"`
}

type DustSample struct {
	BatchID  uuid.UUID `json:"batch_id"`
	MassKG   float64   `json:"mass_kg"`
	Chloride float64   `json:"chloride_percent"`
	Observed time.Time `json:"observed_at"`
}

type BatchSamples struct {
	BatchID uuid.UUID   `json:"batch_id"`
	Gas     *GasSample  `json:"gas,omitempty"`
	Dust    *DustSample `json:"dust,omitempty"`
}

func (b BatchSamples) Complete() bool {
	return b.BatchID != uuid.Nil && b.Gas != nil && b.Dust != nil && b.Gas.BatchID == b.BatchID && b.Dust.BatchID == b.BatchID
}

type Aggregator struct {
	mu      sync.RWMutex
	batches map[uuid.UUID]BatchSamples
}

func NewAggregator() *Aggregator {
	return &Aggregator{batches: make(map[uuid.UUID]BatchSamples)}
}

func (a *Aggregator) UpdateGas(sample GasSample) (BatchSamples, error) {
	if sample.BatchID == uuid.Nil || sample.Chlorine < 0 || sample.Observed.IsZero() {
		return BatchSamples{}, fmt.Errorf("gas chlorine sample is invalid")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	batch := a.batches[sample.BatchID]
	batch.BatchID = sample.BatchID
	batch.Gas = &sample
	a.batches[sample.BatchID] = batch
	return batch, nil
}

func (a *Aggregator) UpdateDust(sample DustSample) (BatchSamples, error) {
	if sample.BatchID == uuid.Nil || sample.MassKG <= 0 || sample.Chloride < 0 || sample.Observed.IsZero() {
		return BatchSamples{}, fmt.Errorf("dust chloride sample is invalid")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	batch := a.batches[sample.BatchID]
	batch.BatchID = sample.BatchID
	batch.Dust = &sample
	a.batches[sample.BatchID] = batch
	return batch, nil
}

func (a *Aggregator) Complete(batchID uuid.UUID) (BatchSamples, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	batch := a.batches[batchID]
	return batch, batch.Complete()
}

func (a *Aggregator) Pending() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	count := 0
	for _, batch := range a.batches {
		if !batch.Complete() {
			count++
		}
	}
	return count
}
