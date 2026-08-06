package liveFeature

import (
	"database/sql"
	"github.com/jmoiron/sqlx/types"
	"time"
)

type LiveFeature struct {
	UUID           string             `json:"id" db:"id"`
	Name           string             `json:"name" db:"name"`
	Category       string             `json:"category" db:"category"`
	Channel        string             `json:"channel" db:"channel"`
	AdditionalInfo types.NullJSONText `json:"additionalInfo" db:"additional_info"`
	CreatedAt      time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time          `json:"updatedAt" db:"updated_at"`
	DeletedAt      sql.NullTime       `json:"deletedAt" db:"deleted_at"`
}
