package paymentModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

type StaticVaFilterRequest struct {
	MerchantID string
	Status     string
	ID         string // search by reference_id or va_number
	BankName   string // search by bank name
	StartDate  time.Time
	EndDate    time.Time
	Sort       string
	SortBy     string
	Page       int
	PerPage    int
}

type StaticVaDetailRequest struct {
	PaymentID  string
	MerchantID string
}

type StaticVaListResponse struct {
	UUID        string    `json:"uuid" db:"uuid"`
	ReferenceID string    `json:"referenceId" db:"reference_id"`
	VaNumber    string    `json:"vaNumber" db:"va_number"`
	VaBank      string    `json:"vaBank" db:"va_bank"`
	VaBankLogo  string    `json:"vaBankLogo" db:"va_bank_logo"`
	VaName      string    `json:"vaName" db:"va_name"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
}

type StaticVaDetailResponse struct {
	UUID                string             `json:"uuid" db:"uuid"`
	ReferenceID         string             `json:"referenceId" db:"reference_id"`
	VaNumber            string             `json:"vaNumber" db:"va_number"`
	VaBank              string             `json:"vaBank" db:"va_bank"`
	VaBankLogo          string             `json:"vaBankLogo" db:"va_bank_logo"`
	VaName              string             `json:"vaName" db:"va_name"`
	VaIssuer            string             `json:"vaIssuer" db:"va_issuer"`
	VaType              string             `json:"vaType" db:"va_type"`
	Status              string             `json:"status" db:"status"`
	CreatedAt           time.Time          `json:"createdAt" db:"created_at"`
	ExpiredAt           *time.Time         `json:"expiredAt,omitempty" db:"expired_at"`
	TotalPaymentCount   int                `json:"totalPaymentAccepted" db:"total_payment_count"`
	TotalAmountValue    string             `json:"-" db:"total_amount_value"`
	TotalAmount         commonModel.Amount `json:"totalAmount"`
	StatementDescriptor *string            `json:"statementDescriptor,omitempty" db:"statement_descriptor"`
}

type StaticVaTransactionItem struct {
	UUID            string             `json:"uuid" db:"uuid"`
	ReferenceID     string             `json:"referenceId" db:"reference_id"`
	AmountValue     string             `json:"-" db:"amount_value"`
	AmountCurrency  string             `json:"-" db:"amount_currency"`
	Status          string             `json:"status" db:"status"`
	CreatedAt       time.Time          `json:"createdAt" db:"created_at"`
	PaymentDate     *time.Time         `json:"paymentDate,omitempty" db:"payment_date"`
	ProcessorRefID  string             `json:"processorReferenceId" db:"processor_reference_id"`
	BankReferenceID string             `json:"bankReferenceId" db:"bank_reference_id"`
	Amount          commonModel.Amount `json:"amount"`
}

type StaticVaTransactionFilterRequest struct {
	PaymentID  string
	MerchantID string
	ID         string    // filter by uuid
	Status     string    // filter by transaction status
	StartDate  time.Time // filter by transaction date
	EndDate    time.Time // filter by transaction date
	Sort       string    // ASC, DESC
	SortBy     string    // createdAt, paymentDate, amount
	Page       int
	PerPage    int
}

type StaticVaUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}

// Default values for filtering
const (
	DefaultStaticVaPage              = 1
	DefaultStaticVaPerPage           = 12
	DefaultStaticVaSort              = "DESC"
	DefaultStaticVaSortBy            = "createdAt"
	DefaultStaticVaTransactionSortBy = "paymentDate"
)

// Validate validates the StaticVaFilterRequest
func (r *StaticVaFilterRequest) Validate() {
	if r.Page < 1 {
		r.Page = DefaultStaticVaPage
	}
	if r.PerPage < 1 {
		r.PerPage = DefaultStaticVaPerPage
	}
	if r.Sort == "" {
		r.Sort = DefaultStaticVaSort
	}
	if r.SortBy == "" {
		r.SortBy = DefaultStaticVaSortBy
	}
}

// Validate validates the StaticVaTransactionFilterRequest
func (r *StaticVaTransactionFilterRequest) Validate() {
	if r.Page < 1 {
		r.Page = DefaultStaticVaPage
	}
	if r.PerPage < 1 {
		r.PerPage = DefaultStaticVaPerPage
	}
	if r.Sort == "" {
		r.Sort = DefaultStaticVaSort
	}
	if r.SortBy == "" {
		r.SortBy = DefaultStaticVaTransactionSortBy
	}
}
