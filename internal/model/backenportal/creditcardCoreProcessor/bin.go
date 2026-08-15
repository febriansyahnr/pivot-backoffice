package creditcardCoreProcessorModel

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Bin struct {
	UUID          uuid.UUID    `json:"uuid" db:"uuid"`
	BinNumber     string       `json:"bin_number" db:"bin_number"`
	CardType      string       `json:"card_type" db:"card_type"`
	CardBrand     string       `json:"card_brand" db:"card_brand"`
	ConsumerType  string       `json:"consumer_type" db:"consumer_type"`
	CardLevel     string       `json:"card_level" db:"card_level"`
	IssuerName    string       `json:"issuer_name" db:"issuer_name"`
	IssuerCountry string       `json:"issuer_country" db:"issuer_country"`
	Currency      string       `json:"currency" db:"currency"`
	Status        string       `json:"status" db:"status"`
	IsBlocked     bool         `json:"is_blocked" db:"is_blocked"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at,omitempty" db:"deleted_at"`
}
