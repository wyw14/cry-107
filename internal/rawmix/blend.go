package rawmix

import (
	"fmt"
	"math"
)

type Component struct {
	Name     string  `json:"name"`
	Fraction float64 `json:"fraction"`
	Moisture float64 `json:"moisture"`
	HeatCost float64 `json:"heat_cost"`
}

type Blend struct {
	Components []Component `json:"components"`
	DryRate    float64     `json:"dry_rate"`
	HeatDemand float64     `json:"heat_demand"`
}

func CalculateBlend(target Target, components []Component) (Blend, error) {
	if len(components) < 2 {
		return Blend{}, fmt.Errorf("raw-mix blend requires at least two components")
	}
	var fraction, weightedMoisture, heat float64
	for _, component := range components {
		if component.Name == "" || component.Fraction < 0 || component.Moisture < 0 || component.Moisture >= 1 {
			return Blend{}, fmt.Errorf("invalid raw-mix component")
		}
		fraction += component.Fraction
		weightedMoisture += component.Fraction * component.Moisture
		heat += component.Fraction * component.HeatCost
	}
	if math.Abs(fraction-1) > 0.001 {
		return Blend{}, fmt.Errorf("raw-mix component fractions total %.4f", fraction)
	}
	dry := target.TonnesPerHour * (1 - weightedMoisture)
	return Blend{Components: append([]Component(nil), components...), DryRate: dry, HeatDemand: dry * heat}, nil
}

func DefaultBlend(target Target) (Blend, error) {
	return CalculateBlend(target, []Component{
		{Name: "limestone", Fraction: 0.82, Moisture: 0.035, HeatCost: 0.78},
		{Name: "clay", Fraction: 0.13, Moisture: 0.09, HeatCost: 0.55},
		{Name: "corrective", Fraction: 0.05, Moisture: 0.02, HeatCost: 0.22},
	})
}
