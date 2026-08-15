package callback_model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CallbackMaster struct {
	UUID        uuid.UUID    `db:"uuid" json:"uuid"`
	Name        string       `db:"name" json:"name"`
	Description string       `db:"description" json:"description"`
	CreatedAt   time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time    `db:"updated_at" json:"updatedAt"`
	DeletedAt   sql.NullTime `db:"deleted_at" json:"deletedAt,omitempty"`
}
