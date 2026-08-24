package kiln

import (
	"fmt"
	"sync"

	"github.com/wyw14/cry-107/internal/lube"
	"github.com/wyw14/cry-107/internal/model"
)

type DrivePermitState struct {
	Identity      model.Identity `json:"identity"`
	Allowed       bool           `json:"allowed"`
	SpeedCeiling  float64        `json:"speed_ceiling_rpm"`
	MissingProofs []string       `json:"missing_proofs"`
}

type DrivePermit struct {
	mu    sync.RWMutex
	state DrivePermitState
}

func NewDrivePermit() *DrivePermit {
	return &DrivePermit{state: DrivePermitState{Allowed: true, SpeedCeiling: 4.8}}
}

func (p *DrivePermit) ApplyHandover(status lube.HandoverStatus) DrivePermitState {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.Identity = status.Identity
	p.state.Allowed = status.Complete
	p.state.MissingProofs = append([]string(nil), status.Missing...)
	if status.Complete {
		p.state.SpeedCeiling = 4.8
	} else {
		p.state.SpeedCeiling = 0.8
	}
	return p.state
}

func (p *DrivePermit) Check(identity model.Identity, speed float64) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if !p.state.Allowed && speed > p.state.SpeedCeiling {
		return fmt.Errorf("kiln drive remains derated: %v", p.state.MissingProofs)
	}
	if p.state.Identity.Generation > 0 && !p.state.Identity.SameGeneration(identity) {
		return fmt.Errorf("kiln drive permit belongs to another handover")
	}
	return nil
}

func (p *DrivePermit) Snapshot() DrivePermitState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}
