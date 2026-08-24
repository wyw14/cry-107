package emission

import "fmt"

type Limits struct {
	COPPM       float64 `json:"co_ppm"`
	NOxPPM      float64 `json:"nox_ppm"`
	DustMG      float64 `json:"dust_mg_m3"`
	ChlorinePPM float64 `json:"chlorine_ppm"`
}

func DefaultLimits() Limits {
	return Limits{COPPM: 1200, NOxPPM: 800, DustMG: 30, ChlorinePPM: 18}
}

func (l Limits) Breaches(sample Sample) []string {
	var values []string
	if sample.CO > l.COPPM {
		values = append(values, fmt.Sprintf("CO %.0f exceeds %.0f", sample.CO, l.COPPM))
	}
	if sample.NOx > l.NOxPPM {
		values = append(values, fmt.Sprintf("NOx %.0f exceeds %.0f", sample.NOx, l.NOxPPM))
	}
	if sample.Dust > l.DustMG {
		values = append(values, fmt.Sprintf("dust %.1f exceeds %.1f", sample.Dust, l.DustMG))
	}
	if sample.Chlorine > l.ChlorinePPM {
		values = append(values, fmt.Sprintf("chlorine %.1f exceeds %.1f", sample.Chlorine, l.ChlorinePPM))
	}
	return values
}
