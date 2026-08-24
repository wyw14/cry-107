package operation

import (
	"fmt"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/model"
)

type Record struct {
	Identity   model.Identity `json:"identity"`
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	StartedAt  time.Time      `json:"started_at"`
	FinishedAt time.Time      `json:"finished_at,omitempty"`
	Detail     string         `json:"detail"`
}

type Manager struct {
	mu          sync.RWMutex
	generations *Generations
	records     map[string]Record
}

func NewManager(generations *Generations) *Manager {
	return &Manager{generations: generations, records: make(map[string]Record)}
}

func (m *Manager) Begin(name string) (Record, error) {
	if name == "" {
		return Record{}, fmt.Errorf("operation name is required")
	}
	identity := m.generations.Begin()
	record := Record{Identity: identity, Name: name, Status: "running", StartedAt: time.Now().UTC()}
	m.mu.Lock()
	m.records[identity.Key()] = record
	m.mu.Unlock()
	return record, nil
}

func (m *Manager) Complete(identity model.Identity, detail string) (Record, error) {
	if err := m.generations.Accept(identity); err != nil {
		return Record{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.records[identity.Key()]
	if !ok {
		return Record{}, fmt.Errorf("operation record not found")
	}
	if record.Status != "running" {
		return Record{}, fmt.Errorf("operation already terminal: %s", record.Status)
	}
	record.Status = "completed"
	record.Detail = detail
	record.FinishedAt = time.Now().UTC()
	m.records[identity.Key()] = record
	return record, nil
}

func (m *Manager) Fail(identity model.Identity, detail string) (Record, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, ok := m.records[identity.Key()]
	if !ok {
		return Record{}, fmt.Errorf("operation record not found")
	}
	record.Status = "failed"
	record.Detail = detail
	record.FinishedAt = time.Now().UTC()
	m.records[identity.Key()] = record
	return record, nil
}

func (m *Manager) List() []Record {
	m.mu.RLock()
	defer m.mu.RUnlock()
	values := make([]Record, 0, len(m.records))
	for _, record := range m.records {
		values = append(values, record)
	}
	return values
}
