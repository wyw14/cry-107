package kiln

import (
	"sort"
	"sync"
	"time"

	"github.com/wyw14/cry-107/internal/shell"
)

type LiningZone struct {
	Segment       int           `json:"segment"`
	Temperature   float64       `json:"temperature_c"`
	HotDuration   time.Duration `json:"hot_duration"`
	DeratePercent float64       `json:"derate_percent"`
}

type LiningMonitor struct {
	mu    sync.RWMutex
	zones map[int]LiningZone
}

func NewLiningMonitor() *LiningMonitor {
	return &LiningMonitor{zones: make(map[int]LiningZone)}
}

func (m *LiningMonitor) ApplyHotspots(spots []shell.Hotspot) []LiningZone {
	m.mu.Lock()
	defer m.mu.Unlock()
	active := make(map[int]struct{}, len(spots))
	for _, spot := range spots {
		zone := LiningZone{Segment: spot.Segment, Temperature: spot.Temperature, HotDuration: spot.Duration}
		if spot.Temperature >= 410 && spot.Duration >= 3*time.Minute {
			zone.DeratePercent = 18
		} else if spot.Temperature >= 380 {
			zone.DeratePercent = 8
		}
		m.zones[spot.Segment] = zone
		active[spot.Segment] = struct{}{}
	}
	for segment := range m.zones {
		if _, ok := active[segment]; !ok {
			delete(m.zones, segment)
		}
	}
	return m.listLocked()
}

func (m *LiningMonitor) Zones() []LiningZone {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.listLocked()
}

func (m *LiningMonitor) listLocked() []LiningZone {
	values := make([]LiningZone, 0, len(m.zones))
	for _, zone := range m.zones {
		values = append(values, zone)
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Segment < values[j].Segment })
	return values
}
