package creditcardCoreProcessorModel

import (
	"time"

	"github.com/google/uuid"
)

type VoidRequest struct {
	MerchantID            string `json:"-"`
	ClientTransactionID   string `json:"client_transaction_id" validate:"required"`
	AcquirerTransactionID string `json:"acquirer_transaction_id,omitempty" validate:"required"`
}

type GetTransactionListRequest struct {
	Limit               int    `query:"limit"`
	Page                int    `query:"page"`
	DateFrom            string `query:"date_from"`
	DateTo              string `query:"date_to"`
	TrxType             string `query:"type"`
	ChargeStatus        string `query:"charge_status"`
	VoidStatus          string `query:"void_status"`
	ClientTransactionID string `query:"client_transaction_id"`
	IssuingBank         string `query:"issuing_bank"`
	CardFingerprint     string `query:"card_fingerprint"`
	PaymentUUID         string `query:"payment_uuid"`
	MerchantID          string `query:"merchant_id"`
	ChargeFrom          string `query:"charge_from"`
	ChargeTo            string `query:"charge_to"`
	RefundFrom          string `query:"refund_from"`
	RefundTo            string `query:"refund_to"`
}

type RefundRequest struct {
	MerchantID              string  `json:"-"`
	ClientTransactionID     string  `json:"client_transaction_id" validate:"required"`
	AcquirerTransactionID   string  `json:"acquirer_transaction_id,omitempty"`
	RefundClientReferenceID string  `json:"refund_client_reference_id,omitempty"`
	Currency                string  `json:"currency,omitempty"`
	Amount                  float64 `json:"amount,omitempty"`
}

type CreateMIDRequest struct {
	Mid                string   `json:"mid" validate:"required"`
	Name               string   `json:"name" validate:"required"`
	Description        string   `json:"description"`
	Type               string   `json:"type"`
	TransactionType    string   `json:"transaction_type"`
	InstallmentType    string   `json:"installment_type"`
	InstallmentTenor   int      `json:"installment_tenor"`
	Processor          string   `json:"processor" validate:"required"`
	PrincipalAvailable []string `json:"principal_available"  validate:"required"`
	IsActive           bool     `json:"is_active"`
	IsDefault          bool     `json:"is_default"`
	BaseURL            string   `json:"base_url" validate:"required"`
	Password           string   `json:"password" validate:"required"`
	Acquirer           string   `json:"acquirer"`
}

type UpdateMIDRequest struct {
	Mid                string   `json:"mid,omitempty"`
	Name               string   `json:"name,omitempty"`
	Description        string   `json:"description,omitempty"`
	Type               string   `json:"type,omitempty"`
	TransactionType    string   `json:"transactionType,omitempty"`
	InstallmentType    string   `json:"installmentType,omitempty"`
	InstallmentTenor   int      `json:"installmentTenor,omitempty"`
	Processor          string   `json:"processor,omitempty"`
	PrincipalAvailable []string `json:"principal_available,omitempty"`
	IsActive           bool     `json:"is_active"`
	IsDefault          bool     `json:"is_default"`
	BaseURL            string   `json:"base_url,omitempty"`
	Password           string   `json:"password,omitempty"`
	Acquirer           string   `json:"acquirer,omitempty"`

	UUID string `json:"-"`
}

type ValidateMIDInstallmentBinsRequest struct {
	MidID string   `json:"mid_id"`
	Bins  []string `json:"bin_numbers" `
}

type CreateMIDMapRequest struct {
	MerchantID uuid.UUID `json:"merchant_id" validate:"required"`
	MidID      uuid.UUID `json:"mid_id"`
	Mid        string    `json:"mid"`
	IsActive   bool      `json:"is_active"`
	Priority   int       `json:"priority"`
}

type UpdateMIDMapPriorityRequest struct {
	IsActive bool `json:"is_active"`
	Priority int  `json:"priority"`

	MidMapID uuid.UUID `json:"-"`
}

type GetMIDListRequest struct {
	Limit           int `json:"limit"`
	Page            int `json:"page"`
	Mid             string
	Acquirer        string
	Name            string
	Type            string
	TransactionType string
	InstallmentType string
	IsDefault       *bool
	IsActive        *bool
}

type GetMIDMapListRequest struct {
	Limit      int    `json:"limit"`
	Page       int    `json:"page"`
	MerchantId string `json:"merchant_id"`
}

type FindMIDMapByMerchantRequest struct {
	MerchantID string `json:"merchant_id"`
	MidID      string `json:"mid_Id"`
}

type BlockCardRequest struct {
	CardUUID    string    `json:"cardUuid"`
	IsBlocked   bool      `json:"isBlocked"`
	BlockedTo   time.Time `json:"blockedTo,omitempty"`
	BlockReason string    `json:"blockReason,omitempty"`
}

type CaptureRequest struct {
	MerchantID             string  `json:"-"`
	ClientTransactionID    string  `json:"client_transaction_id" validate:"required"`
	AcquirerTransactionID  string  `json:"acquirer_transaction_id"`
	ReleaseRemainingAmount bool    `json:"release_remaining_amount"`
	Currency               string  `json:"currency"`
	Amount                 float64 `json:"amount"`
	CapturedAmount         float64 `json:"captured_amount,omitempty"`
}

type AuthenticationRequest struct {
	MerchantID       string
	PaymentID        string
	EncryptedPayload string
}
