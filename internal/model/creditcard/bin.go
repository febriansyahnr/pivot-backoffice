package card

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Bin struct {
	UUID          uuid.UUID    `json:"uuid"`
	BinNumber     string       `json:"binNumber"`
	CardType      string       `json:"cardType"`
	CardBrand     string       `json:"cardBrand"`
	ConsumerType  string       `json:"consumerType"`
	CardLevel     string       `json:"cardLevel"`
	IssuerName    string       `json:"issuerName"`
	IssuerCountry string       `json:"issuerCountry"`
	Currency      string       `json:"currency"`
	Status        string       `json:"status"`
	IsBlocked     bool         `json:"isBlocked"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
	DeletedAt     sql.NullTime `json:"deletedAt,omitempty"`
}
