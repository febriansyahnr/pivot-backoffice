package reportingModel

import (
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

// BalanceHistory represents the report_balance_histories table row
type BalanceHistory struct {
	TransactionID     string                        `json:"transactionId" db:"transaction_id"`
	MerchantID        string                        `json:"merchantId" db:"merchant_id"`
	ReferenceID       string                        `json:"referenceId" db:"reference_id"`
	Type              string                        `json:"type" db:"type"` // Enum for usecase type
	BalanceType       string                        `json:"balanceType" db:"balance_type"`
	Channel           string                        `json:"channel" db:"channel"`
	TransactionType   string                        `json:"transactionType" db:"transaction_type"` // Enum for sub usecase type
	Currency          string                        `json:"currency" db:"currency"`
	Amount            decimal.Decimal               `json:"amount" db:"amount"`
	Fee               decimal.Decimal               `json:"fee" db:"fee"`
	Remarks           string                        `json:"remarks" db:"remarks"`
	Status            string                        `json:"status" db:"status"`
	ReasonType        *string                       `json:"reasonType" db:"reason_type"`
	ReasonDescription *string                       `json:"reasonDescription" db:"reason_description"`
	SettlementModel   string                        `json:"settlementModel" db:"settlement_model"`
	SettlementStatus  string                        `json:"settlementStatus" db:"settlement_status"`
	SettlementAt      time.Time                     `json:"settlementAt" db:"settlement_at"`
	RawAdditionalInfo types.NullJSONText            `json:"-" db:"additional_info"`
	AdditionalInfo    *BalanceHistoryAdditionalInfo `json:"additionalInfo" db:"-"`
	CreatedAt         time.Time                     `json:"createdAt" db:"created_at"`
	StatusUpdatedAt   time.Time                     `json:"statusUpdatedAt" db:"status_updated_at"`
	SourceID          string                        `json:"sourceId" db:"source_id"`
	SourceAccountID   string                        `json:"sourceAccountId" db:"source_account_id"`
	SourceCreatedAt   *time.Time                    `json:"sourceCreatedAt" db:"source_created_at"`
	SourceCreatedBy   string                        `json:"sourceCreatedBy" db:"source_created_by"`
	IsDeleted         bool                          `json:"-" db:"_is_deleted"`
	DeletedAt         *time.Time                    `json:"-" db:"_deleted_at"`
	IngestedAt        time.Time                     `json:"-" db:"_ingested_at"`
}

type BalanceHistoryAdditionalInfo struct {
	BankReferenceNo      string `json:"bankReferenceNo,omitempty"`
	BeneficiaryAccountNo string `json:"beneficiaryAccountNo,omitempty"`
	BeneficiaryName      string `json:"beneficiaryName,omitempty"`
	BeneficiaryBankName  string `json:"beneficiaryBankName,omitempty"`
}
