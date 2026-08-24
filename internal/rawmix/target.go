package rawmix

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type Target struct {
	Identity       model.Identity `json:"identity"`
	TonnesPerHour  float64        `json:"tonnes_per_hour"`
	LimeSaturation float64        `json:"lime_saturation"`
	SilicaModulus  float64        `json:"silica_modulus"`
	CreatedAt      time.Time      `json:"created_at"`
}

type Planner struct {
	mu      sync.RWMutex
	current Target
}

func NewPlanner() *Planner {
	return &Planner{}
}

func (p *Planner) Set(identity model.Identity, tonnes, lime, silica float64) (Target, error) {
	if !identity.Valid() {
		return Target{}, fmt.Errorf("raw-mix target needs a valid operation")
	}
	if tonnes < 0 || tonnes > 500 {
		return Target{}, fmt.Errorf("raw-mix feed target %.1f is outside operating range", tonnes)
	}
	if lime < 0.8 || lime > 1.2 || silica < 1.5 || silica > 3.5 {
		return Target{}, fmt.Errorf("raw-mix chemistry target is outside operating range")
	}
	target := Target{
		Identity: identity, TonnesPerHour: tonnes,
		LimeSaturation: lime, SilicaModulus: silica, CreatedAt: time.Now().UTC(),
	}
	p.mu.Lock()
	p.current = target
	p.mu.Unlock()
	return target, nil
}

func (p *Planner) Current() Target {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.current
}

func (p *Planner) KilnSpeedRPM(fillFraction float64) (float64, error) {
	p.mu.RLock()
	target := p.current
	p.mu.RUnlock()
	if target.Identity.Generation == 0 {
		return 0, fmt.Errorf("raw-mix target has not been established")
	}
	if fillFraction <= 0 || fillFraction > 0.3 {
		return 0, fmt.Errorf("kiln fill fraction is invalid")
	}
	rpm := target.TonnesPerHour / (110 * fillFraction)
	return math.Max(0.4, math.Min(4.8, rpm)), nil
}
