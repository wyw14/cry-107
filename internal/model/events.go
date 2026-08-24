package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID         uuid.UUID       `json:"id"`
	Kind       string          `json:"kind"`
	Operation  uuid.UUID       `json:"operation_id"`
	Generation uint64          `json:"generation"`
	OccurredAt time.Time       `json:"occurred_at"`
	Payload    json.RawMessage `json:"payload"`
}

func NewEvent(kind string, identity Identity, payload any) (Event, error) {
	if kind == "" {
		return Event{}, fmt.Errorf("event kind is required")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return Event{}, fmt.Errorf("encode %s event: %w", kind, err)
	}
	return Event{
		ID:         uuid.New(),
		Kind:       kind,
		Operation:  identity.Operation,
		Generation: identity.Generation,
		OccurredAt: time.Now().UTC(),
		Payload:    body,
	}, nil
}

func (e Event) Validate() error {
	if e.ID == uuid.Nil || e.Operation == uuid.Nil {
		return fmt.Errorf("event identity is incomplete")
	}
	if e.Kind == "" || e.Generation == 0 || e.OccurredAt.IsZero() {
		return fmt.Errorf("event metadata is incomplete")
	}
	if !json.Valid(e.Payload) {
		return fmt.Errorf("event payload is not valid JSON")
	}
	return nil
}

func DecodePayload[T any](event Event) (T, error) {
	var value T
	if err := json.Unmarshal(event.Payload, &value); err != nil {
		return value, fmt.Errorf("decode %s payload: %w", event.Kind, err)
	}
	return value, nil
}
