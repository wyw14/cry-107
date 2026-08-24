package verifycase

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/bypass"
	"github.com/wyw14/cry-107/internal/sampler"
)

func TestChlorineEnrichmentUsesOneResidenceBatch(t *testing.T) {
	aggregator := sampler.NewAggregator()
	service := bypass.NewEnrichment(aggregator)
	oldBatch := uuid.New()
	newBatch := uuid.New()
	if _, err := aggregator.UpdateDust(sampler.DustSample{BatchID: oldBatch, MassKG: 18, Chloride: 2.4, Observed: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if _, err := aggregator.UpdateGas(sampler.GasSample{BatchID: newBatch, Chlorine: 14, Observed: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	if result, err := service.Calculate(newBatch); err == nil {
		t.Fatalf("cross-batch measurements produced enrichment: %+v", result)
	}
	if _, err := aggregator.UpdateDust(sampler.DustSample{BatchID: newBatch, MassKG: 22, Chloride: 1.7, Observed: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	result, err := service.Calculate(newBatch)
	if err != nil || result.BatchID != newBatch {
		t.Fatalf("complete residence batch was not calculated: %+v %v", result, err)
	}
}
