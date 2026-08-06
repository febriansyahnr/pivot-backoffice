package callback_model

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Callback struct {
	UUID             uuid.UUID    `db:"uuid" json:"uuid"`
	CallbackMasterID uuid.UUID    `db:"callback_master_id" json:"callbackMasterId"`
	MerchantID       uuid.UUID    `db:"merchant_id" json:"merchantId"`
	BaseURL          *string      `db:"base_url" json:"baseUrl"`
	URL              string       `db:"url" json:"url"`
	Description      string       `db:"description" json:"description"`
	CreatedAt        time.Time    `db:"created_at" json:"createdAt"`
	UpdatedAt        time.Time    `db:"updated_at" json:"updatedAt"`
	DeletedAt        sql.NullTime `db:"deleted_at" json:"deletedAt,omitempty"`
}
