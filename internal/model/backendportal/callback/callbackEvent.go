package callback_model

import (
	"time"

	"github.com/google/uuid"
)

type CallbackEvent struct {
	UUID       uuid.UUID `db:"uuid" json:"uuid"`
	// Event maps to "value" in JSON for frontend dropdown compatibility
	Event      string    `db:"event" json:"value"`
	Label      string    `db:"label" json:"label"`
	EventGroup string    `db:"event_group" json:"group"`
	IsActive   bool      `db:"is_active" json:"-"`
	CreatedAt  time.Time `db:"created_at" json:"-"`
	UpdatedAt  time.Time `db:"updated_at" json:"-"`
}
