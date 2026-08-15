package merchant

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx/types"
)

type BoardOfDirector struct {
	Id             string         `db:"id"`
	MerchantId     string         `db:"merchant_id"`
	Position       string         `db:"position"` // Oneof: Director or Commissioner
	Name           string         `db:"name"`
	Shares         *float64       `db:"shares"`
	IdentityNumber string         `db:"identity_number"`
	IdentityFile   types.JSONText `db:"identity_file"`
	Hash           string         `db:"hash"`
	PositionLong   string         `db:"position_long"`
	Status         string         `db:"status"`
	CreatedBy      string         `db:"created_by"`
	CreatedAt      time.Time      `db:"created_at"`
	ApprovedBy     string         `db:"approved_by"`
	ApprovedAt     sql.NullTime   `db:"approved_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      sql.NullTime   `db:"deleted_at"`
}
