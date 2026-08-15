package refundModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type RefundCallback struct {
	ID                string             `json:"id"`
	ClientReferenceID string             `json:"clientReferenceId"`
	MerchantID        string             `json:"merchantId"`
	PaymentID         string             `json:"paymentId"`
	PaymentChargeID   string             `json:"paymentChargeId"`
	Amount            commonModel.Amount `json:"amount"`
	Status            string             `json:"status"`
	Method            string             `json:"method"`
	DestinationType   string             `json:"destinationType"`
	TransactionTime   time.Time          `json:"transactionTime"`
	Reason            string             `json:"reason,omitempty"`
	Description       string             `json:"description,omitempty"`
	Metadata          interface{}        `json:"metadata,omitempty"`
}

type RefundStatusChanged struct {
	RefundID   string
	OldStatus  string
	NewStatus  string
	MerchantID string
	UpdatedAt  time.Time
}
