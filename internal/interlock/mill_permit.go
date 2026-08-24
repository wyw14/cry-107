package interlock

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type MillPermitState struct {
	SessionID uuid.UUID `json:"session_id"`
	HotAir    bool      `json:"hot_air"`
	CoalFeed  bool      `json:"coal_feed"`
	Reason    string    `json:"reason"`
	Changed   time.Time `json:"changed_at"`
}

type MillPermit struct {
	mu    sync.RWMutex
	state MillPermitState
}

func NewMillPermit() *MillPermit {
	return &MillPermit{state: MillPermitState{Reason: "inerting not complete", Changed: time.Now().UTC()}}
}

func (p *MillPermit) InertingStarted(session uuid.UUID) {
	p.mu.Lock()
	p.state = MillPermitState{SessionID: session, Reason: "nitrogen purge active", Changed: time.Now().UTC()}
	p.mu.Unlock()
}

func (p *MillPermit) InertingCompleted(session uuid.UUID) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if session == uuid.Nil || session != p.state.SessionID {
		return false
	}
	p.state.HotAir = true
	p.state.CoalFeed = true
	p.state.Reason = "inerting volume and oxygen window proven"
	p.state.Changed = time.Now().UTC()
	return true
}

// InertingRevoked withdraws the hot-air and coal-feed authorisation when the
// inerting window breaks (oxygen above the limit or purge volume no longer
// met). The permit stays bound to the session so the inerting state machine
// can complete again once the window is re-proven; only Trip tears the session
// down for a terminal interlock.
func (p *MillPermit) InertingRevoked(session uuid.UUID, reason string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if session == uuid.Nil || session != p.state.SessionID {
		return
	}
	p.state.HotAir = false
	p.state.CoalFeed = false
	p.state.Reason = reason
	p.state.Changed = time.Now().UTC()
}

func (p *MillPermit) Trip(reason string) MillPermitState {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state.HotAir = false
	p.state.CoalFeed = false
	p.state.Reason = reason
	p.state.Changed = time.Now().UTC()
	return p.state
}

func (p *MillPermit) Snapshot() MillPermitState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}
