package statusHistoryModel

import (
	"time"

	"github.com/jmoiron/sqlx/types"
)

type StatusHistory struct {
	ID            string                 `json:"id" db:"id"`
	ReferenceType string                 `json:"referenceType" db:"reference_type"`
	ReferenceID   string                 `json:"referenceId" db:"reference_id"`
	Status        string                 `json:"status" db:"status"`
	Metadata      types.NullJSONText     `json:"-" db:"metadata"`
	MetadataObj   *StatusHistoryMetadata `json:"metadata" db:"-"`
	CreatedAt     time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time              `json:"updatedAt" db:"updated_at"`
}
