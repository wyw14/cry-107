package shell

import (
	"fmt"
	"math"
	"sync"
	"time"
)

type Calibration struct {
	Revision  uint64    `json:"revision"`
	OffsetDeg float64   `json:"offset_deg"`
	ChangedAt time.Time `json:"changed_at"`
}

func (c Calibration) Segment(scannerAngle float64) int {
	physical := math.Mod(scannerAngle+c.OffsetDeg, 360)
	if physical < 0 {
		physical += 360
	}
	return int(physical / 30)
}

type Scanner struct {
	mu          sync.RWMutex
	calibration Calibration
	tracker     *HotspotTracker
}

func NewScanner(tracker *HotspotTracker) *Scanner {
	return &Scanner{tracker: tracker, calibration: Calibration{Revision: 1, ChangedAt: time.Now().UTC()}}
}

func (s *Scanner) ApplyCorrection(offset float64) (Calibration, error) {
	if offset <= -180 || offset >= 180 {
		return Calibration{}, fmt.Errorf("scanner phase correction is outside range")
	}
	s.mu.Lock()
	previous := s.calibration
	next := Calibration{Revision: previous.Revision + 1, OffsetDeg: offset, ChangedAt: time.Now().UTC()}
	s.calibration = next
	s.mu.Unlock()
	s.tracker.Rebind(previous, next)
	return next, nil
}

func (s *Scanner) Observe(angle, temperature float64, duration time.Duration) Hotspot {
	s.mu.RLock()
	calibration := s.calibration
	s.mu.RUnlock()
	return s.tracker.Observe(calibration, angle, temperature, duration)
}

func (s *Scanner) Calibration() Calibration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.calibration
}

func (s *Scanner) Hotspots() []Hotspot {
	return s.tracker.List()
}
