package vccSettlement

import (
	"database/sql"
	"encoding/json"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type VccSettlement struct {
	UUID                    string          `json:"uuid" db:"uuid"`
	RcnId                   string          `json:"rcnId" db:"rcn_id"`
	AcquirerReferenceNumber string          `json:"acquirerReferenceNumber" db:"acquirer_reference_number"`
	Status                  string          `json:"status" db:"status"`
	ReferenceNo             string          `json:"referenceNo" db:"reference_no"`
	AuthorizationNo         string          `json:"authorizationNo" db:"authorization_no"`
	PostingDate             time.Time       `json:"postingDate" db:"posting_date"`
	BillingCycle            int             `json:"billingCycle" db:"billing_cycle"`
	SourceAmount            json.RawMessage `json:"sourceAmount" db:"source_amount"`
	BillingAmount           json.RawMessage `json:"billingAmount" db:"billing_amount"`
	TransactionDate         time.Time       `json:"transactionDate" db:"transaction_date"`
	SettlementDate          time.Time       `json:"settlementDate" db:"settlement_date"`
	MerchantName            string          `json:"merchantName" db:"merchant_name"`
	MerchantCountry         string          `json:"merchantCountry" db:"merchant_country"`
	MerchantCategory        string          `json:"merchantCategory" db:"merchant_category"`
	CreatedAt               time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt               time.Time       `json:"updatedAt" db:"updated_at"`
	DeletedAt               sql.NullTime    `json:"deletedAt" db:"deleted_at"`

	BillingAmountObj commonModel.Amount
	SourceAmountObj  commonModel.Amount
}
