package refundModel

import (
	"github.com/jmoiron/sqlx/types"
	"time"
)

type Refund struct {
	UUID              string             `json:"uuid" db:"uuid"`
	MerchantID        string             `json:"merchantId" db:"merchant_id"`
	ClientReferenceID string             `json:"clientReferenceId" db:"client_reference_id"`
	PaymentID         string             `json:"paymentId" db:"payment_id"`
	PaymentChargeID   string             `json:"paymentChargeId" db:"payment_charge_id"`
	Currency          string             `json:"currency" db:"currency"`
	Amount            float64            `json:"amount" db:"amount"`
	Status            string             `json:"status" db:"status"`
	Reason            string             `json:"reason" db:"reason"`
	Description       string             `json:"description,omitempty" db:"description"`
	DestinationType   string             `json:"destinationType" db:"destination_type"`
	Method            string             `json:"method" db:"method"`
	CreatedAt         time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time          `json:"updatedAt" db:"updated_at"`
	Metadata          types.NullJSONText `json:"-" db:"metadata"`

	MetadataObj MetadataObj `json:"metadata,omitempty" db:"-"`
}
