package shell

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Hotspot struct {
	ID           uuid.UUID     `json:"id"`
	ScannerAngle float64       `json:"scanner_angle"`
	Segment      int           `json:"segment"`
	Revision     uint64        `json:"revision"`
	Temperature  float64       `json:"temperature_c"`
	Duration     time.Duration `json:"duration"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type HotspotTracker struct {
	mu     sync.RWMutex
	byID   map[uuid.UUID]Hotspot
	byZone map[int][]uuid.UUID
}

func NewHotspotTracker() *HotspotTracker {
	return &HotspotTracker{byID: make(map[uuid.UUID]Hotspot), byZone: make(map[int][]uuid.UUID)}
}

func (t *HotspotTracker) Observe(calibration Calibration, angle, temperature float64, duration time.Duration) Hotspot {
	t.mu.Lock()
	defer t.mu.Unlock()
	segment := calibration.Segment(angle)
	spot := Hotspot{
		ID: uuid.New(), ScannerAngle: angle, Segment: segment, Revision: calibration.Revision,
		Temperature: temperature, Duration: duration, UpdatedAt: time.Now().UTC(),
	}
	t.byID[spot.ID] = spot
	t.byZone[segment] = append(t.byZone[segment], spot.ID)
	return spot
}

func (t *HotspotTracker) Rebind(previous, next Calibration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if previous.Revision == next.Revision {
		return
	}
	t.byZone = make(map[int][]uuid.UUID)
	for id, spot := range t.byID {
		spot.Segment = next.Segment(spot.ScannerAngle)
		spot.Revision = next.Revision
		spot.UpdatedAt = time.Now().UTC()
		t.byID[id] = spot
		t.byZone[spot.Segment] = append(t.byZone[spot.Segment], id)
	}
}

func (t *HotspotTracker) List() []Hotspot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	values := make([]Hotspot, 0, len(t.byID))
	for _, spot := range t.byID {
		values = append(values, spot)
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Segment == values[j].Segment {
			return values[i].UpdatedAt.Before(values[j].UpdatedAt)
		}
		return values[i].Segment < values[j].Segment
	})
	return values
}

func (t *HotspotTracker) Zone(segment int) []Hotspot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	ids := t.byZone[segment]
	values := make([]Hotspot, 0, len(ids))
	for _, id := range ids {
		values = append(values, t.byID[id])
	}
	return values
}
