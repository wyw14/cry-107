package interlock

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/wyw14/cry-107/internal/model"
)

type Incident struct {
	ID         uuid.UUID      `json:"id"`
	Identity   model.Identity `json:"identity"`
	Source     string         `json:"source"`
	Severity   string         `json:"severity"`
	Reason     string         `json:"reason"`
	Terminal   bool           `json:"terminal"`
	OccurredAt time.Time      `json:"occurred_at"`
}

type Manager struct {
	mu        sync.RWMutex
	incidents map[uuid.UUID]Incident
	terminal  map[uint64]uuid.UUID
}

func NewManager() *Manager {
	return &Manager{incidents: make(map[uuid.UUID]Incident), terminal: make(map[uint64]uuid.UUID)}
}

func (m *Manager) Raise(identity model.Identity, source, severity, reason string, terminal bool) (Incident, error) {
	if !identity.Valid() || source == "" || reason == "" {
		return Incident{}, fmt.Errorf("incident identity, source and reason are required")
	}
	incident := Incident{
		ID: uuid.New(), Identity: identity, Source: source, Severity: severity,
		Reason: reason, Terminal: terminal, OccurredAt: time.Now().UTC(),
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if terminal {
		if existing := m.terminal[identity.Generation]; existing != uuid.Nil {
			return m.incidents[existing], nil
		}
		m.terminal[identity.Generation] = incident.ID
	}
	m.incidents[incident.ID] = incident
	return incident, nil
}

func (m *Manager) Terminal(identity model.Identity) (Incident, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	id := m.terminal[identity.Generation]
	incident, ok := m.incidents[id]
	return incident, ok
}

func (m *Manager) List() []Incident {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]Incident, 0, len(m.incidents))
	for _, incident := range m.incidents {
		values = append(values, incident)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].OccurredAt.After(values[j].OccurredAt) })
	return values
}
