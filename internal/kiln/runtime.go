package kiln

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type RuntimeState struct {
	Process     model.ProcessState `json:"process"`
	SpeedRPM    float64            `json:"speed_rpm"`
	FeedTPH     float64            `json:"feed_tph"`
	LoadPercent float64            `json:"load_percent"`
	DraftPascal float64            `json:"draft_pascal"`
	ClinkerTemp float64            `json:"clinker_temp_c"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

type Runtime struct {
	mu    sync.RWMutex
	state RuntimeState
}

func NewRuntime() *Runtime {
	return &Runtime{state: RuntimeState{
		Process: model.NewProcessState(), SpeedRPM: 1.8,
		LoadPercent: 70, DraftPascal: -45, ClinkerTemp: 118,
		UpdatedAt: time.Now().UTC(),
	}}
}

func (r *Runtime) Transition(next model.KilnPhase, reason string) (RuntimeState, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	process, err := r.state.Process.Transition(next, reason)
	if err != nil {
		return r.state, err
	}
	r.state.Process = process
	r.state.UpdatedAt = time.Now().UTC()
	return r.state, nil
}

func (r *Runtime) ApplyTargets(speed, feed, load float64, identity model.Identity) (RuntimeState, error) {
	if !identity.Valid() {
		return RuntimeState{}, fmt.Errorf("kiln target operation is invalid")
	}
	if speed < 0 || speed > 5 || feed < 0 || feed > 500 || load < 0 || load > 100 {
		return RuntimeState{}, fmt.Errorf("kiln target is outside operating envelope")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.SpeedRPM = speed
	r.state.FeedTPH = feed
	r.state.LoadPercent = load
	r.state.UpdatedAt = time.Now().UTC()
	return r.state, nil
}

func (r *Runtime) Observe(draft, clinkerTemperature float64) RuntimeState {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.DraftPascal = draft
	r.state.ClinkerTemp = clinkerTemperature
	r.state.UpdatedAt = time.Now().UTC()
	return r.state
}

func (r *Runtime) Snapshot() RuntimeState {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.state
}
