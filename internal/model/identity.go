package model

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Identity ties state changes to one equipment operation generation.
type Identity struct {
	RunID      uuid.UUID `json:"run_id"`
	Operation  uuid.UUID `json:"operation_id"`
	Generation uint64    `json:"generation"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewIdentity(generation uint64) Identity {
	return Identity{
		RunID:      uuid.New(),
		Operation:  uuid.New(),
		Generation: generation,
		CreatedAt:  time.Now().UTC(),
	}
}

func (i Identity) Valid() bool {
	return i.RunID != uuid.Nil && i.Operation != uuid.Nil && i.Generation > 0
}

func (i Identity) Key() string {
	return fmt.Sprintf("%s/%s/%d", i.RunID, i.Operation, i.Generation)
}

func (i Identity) SameGeneration(other Identity) bool {
	return i.RunID == other.RunID && i.Generation == other.Generation
}

func (i Identity) Next() Identity {
	next := i
	next.Operation = uuid.New()
	next.Generation++
	next.CreatedAt = time.Now().UTC()
	return next
}
