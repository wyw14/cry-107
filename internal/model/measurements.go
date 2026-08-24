package model

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

type Measurement struct {
	SensorID  string    `json:"sensor_id"`
	BatchID   uuid.UUID `json:"batch_id"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Revision  uint64    `json:"revision"`
	Observed  time.Time `json:"observed_at"`
	QualityOK bool      `json:"quality_ok"`
}

func (m Measurement) Validate() error {
	if m.SensorID == "" || m.Unit == "" {
		return fmt.Errorf("measurement sensor and unit are required")
	}
	if m.BatchID == uuid.Nil || m.Revision == 0 || m.Observed.IsZero() {
		return fmt.Errorf("measurement identity is incomplete")
	}
	if math.IsNaN(m.Value) || math.IsInf(m.Value, 0) {
		return fmt.Errorf("measurement value must be finite")
	}
	return nil
}

type Proof struct {
	Name       string    `json:"name"`
	Operation  uuid.UUID `json:"operation_id"`
	Generation uint64    `json:"generation"`
	Satisfied  bool      `json:"satisfied"`
	Detail     string    `json:"detail"`
	ObservedAt time.Time `json:"observed_at"`
}

func (p Proof) Matches(identity Identity) bool {
	return p.Operation == identity.Operation && p.Generation == identity.Generation
}

func (p Proof) Ready(identity Identity) bool {
	return p.Matches(identity) && p.Satisfied && !p.ObservedAt.IsZero()
}
