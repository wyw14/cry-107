package journal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type SnapshotStore struct {
	mu   sync.Mutex
	path string
}

func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

func (s *SnapshotStore) Save(value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal snapshot: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return fmt.Errorf("create snapshot directory: %w", err)
	}
	temporary := s.path + ".next"
	if err := os.WriteFile(temporary, append(body, '\n'), 0o644); err != nil {
		return fmt.Errorf("write snapshot candidate: %w", err)
	}
	if err := os.Rename(temporary, s.path); err != nil {
		return fmt.Errorf("commit snapshot: %w", err)
	}
	return nil
}

func (s *SnapshotStore) Load(target any) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, err := os.ReadFile(s.path)
	if errorsIsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read snapshot: %w", err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return false, fmt.Errorf("decode snapshot: %w", err)
	}
	return true, nil
}

func errorsIsNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}
