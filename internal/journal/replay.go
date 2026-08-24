package journal

import (
	"fmt"
	"sort"

	"github.com/wyw14/cry-107/internal/model"
)

type Reducer interface {
	Apply(model.Event) error
}

type ReplayReport struct {
	Applied          int      `json:"applied"`
	IgnoredDuplicate int      `json:"ignored_duplicate"`
	LastGeneration   uint64   `json:"last_generation"`
	EventKinds       []string `json:"event_kinds"`
}

func Replay(events []model.Event, reducer Reducer) (ReplayReport, error) {
	ordered := append([]model.Event(nil), events...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].OccurredAt.Equal(ordered[j].OccurredAt) {
			return ordered[i].ID.String() < ordered[j].ID.String()
		}
		return ordered[i].OccurredAt.Before(ordered[j].OccurredAt)
	})
	seen := make(map[string]struct{}, len(ordered))
	report := ReplayReport{}
	for _, event := range ordered {
		if _, ok := seen[event.ID.String()]; ok {
			report.IgnoredDuplicate++
			continue
		}
		if err := event.Validate(); err != nil {
			return report, fmt.Errorf("replay validation: %w", err)
		}
		if err := reducer.Apply(event); err != nil {
			return report, fmt.Errorf("apply %s event: %w", event.Kind, err)
		}
		seen[event.ID.String()] = struct{}{}
		report.Applied++
		report.EventKinds = append(report.EventKinds, event.Kind)
		if event.Generation > report.LastGeneration {
			report.LastGeneration = event.Generation
		}
	}
	return report, nil
}
