package reconciliation

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/shopspring/decimal"

	"github.com/jmoiron/sqlx/types"
)

type Transaction struct {
	TransactionDate      time.Time
	Reference            string
	Reference2           string
	Amount               decimal.Decimal
	Bank                 string
	Channel              string
	Merchant             string
	Status               string
	Reason               string
	ReconTime            string
	Order                int
	TransactionReference string
	TransactionType      string
}

type TransactionCheckResponse struct {
	Status string
	Reason string
}

type ReconTransactionModel struct {
	UUID                     string             `db:"uuid"`
	Type                     string             `db:"type"`
	ReferenceID              sql.NullString     `db:"reference_id"`
	MerchantName             string             `db:"merchant_name"`
	Processor                string             `db:"processor_reference"`
	ProcessorID              string             `db:"processor_reference_id"`
	Reference                sql.NullString     `db:"reference"`
	Channel                  string             `db:"channel"`
	Status                   string             `db:"status"`
	ReasonType               sql.NullString     `db:"reason_type"`
	ReasonDesc               sql.NullString     `db:"reason_description"`
	AdditionalInfo           types.NullJSONText `db:"additional_info"`
	TransactionDate          time.Time          `db:"transaction_timestamp"`
	Amount                   decimal.Decimal    `db:"amount"`
	PaymentType              string             `db:"payment_type"`
	ProcessorReferenceNumber string             `db:"processor_reference_number"`
}

type ReconTransactionQuery struct {
	Amount          decimal.Decimal `db:"amount"`
	ReferenceID     string          `db:"reference_id"`
	TransactionDate time.Time       `db:"transaction_timestamp"`
	// Reference like PAYMENT, DISBURSEMENT, etc.
	Reference string `db:"reference"`
	// TransactionType like PAYMENT, DISBURSEMENT, etc.
	TransactionType string `db:"transaction_type"`
	// Channel like CREDIT_CARD, VIRTUAL_ACCOUNT, QRIS
	Channel          string        `db:"channel"`
	WithTimeDuration bool          `db:"-"`
	Duration         time.Duration `db:"-"`
	SettlementModel  string        `db:"-"`

	// Scanning Tolerance
	StartUpdatedAt time.Time `db:"-"`
	EndUpdatedAt   time.Time `db:"-"`
}

type BulkUpatedStatus struct {
	StartTime    time.Time
	EndTime      time.Time
	Status       string
	TrxReference string
	TrxType      string

	ScanningToleranceInDays int
}

type ReconDetail struct {
	Status   string `json:"status,omitempty"`
	Reason   string `json:"reason,omitempty"`
	DateTime string `json:"datetime,omitempty"`
	Amount   string `json:"amount,omitempty"`
}

func (r *ReconDetail) Validate() error {
	if r.Status == "" {
		return errors.New("status is required")
	}

	switch {
	case strings.EqualFold(r.Status, constant.ReconStatusReview):
		r.Status = constant.ReconStatusReview
	case strings.EqualFold(r.Status, constant.ReconStatusSuccess):
		r.Status = constant.ReconStatusSuccess
	default:
		return constant.ErrInvalidStatus
	}

	return nil
}

type PaymentTotalAmountQuery struct {
	ReferenceIDs []string
	Channel      string
	StartTime    time.Time
	EndTime      time.Time
}

func (p *PaymentTotalAmountQuery) GetReferenceIDQuery() string {
	return strings.Join(p.ReferenceIDs, "','")
}

type PaymentTotalAmountResult map[string]decimal.Decimal

func (p *PaymentTotalAmountResult) Add(reference string, amount decimal.Decimal) {
	if _, ok := (*p)[reference]; !ok {
		(*p)[reference] = amount
	} else {
		(*p)[reference] = (*p)[reference].Add(amount)
	}
}

func (p *PaymentTotalAmountResult) GetTotalAmount(reference string) decimal.Decimal {
	if _, ok := (*p)[reference]; !ok {
		return decimal.Zero
	}
	return (*p)[reference]
}
