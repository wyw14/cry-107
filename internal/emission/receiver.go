package emission

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type Sample struct {
	Identity   model.Identity `json:"identity"`
	CO         float64        `json:"co_ppm"`
	NOx        float64        `json:"nox_ppm"`
	Dust       float64        `json:"dust_mg_m3"`
	Chlorine   float64        `json:"chlorine_ppm"`
	ObservedAt time.Time      `json:"observed_at"`
}

type Receiver struct {
	mu       sync.RWMutex
	limits   Limits
	latest   Sample
	breaches []string
}

func NewReceiver(limits Limits) *Receiver {
	return &Receiver{limits: limits}
}

func (r *Receiver) Receive(sample Sample) ([]string, error) {
	if !sample.Identity.Valid() || sample.ObservedAt.IsZero() {
		return nil, fmt.Errorf("emission sample identity and timestamp are required")
	}
	if sample.CO < 0 || sample.NOx < 0 || sample.Dust < 0 || sample.Chlorine < 0 {
		return nil, fmt.Errorf("emission sample cannot contain negative values")
	}
	breaches := r.limits.Breaches(sample)
	r.mu.Lock()
	r.latest = sample
	r.breaches = append([]string(nil), breaches...)
	r.mu.Unlock()
	return breaches, nil
}

func (r *Receiver) Summary() (Sample, string) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.latest, strings.Join(r.breaches, "; ")
}
