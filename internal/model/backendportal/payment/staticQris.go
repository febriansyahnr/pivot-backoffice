package paymentModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type StaticQrisFilterRequest struct {
	MerchantID      string
	Status          string
	ID              string
	PaymentMethodID string
	StartDate       time.Time
	EndDate         time.Time
	Sort            string
	SortBy          string
	Page            int
	PerPage         int
}
type StaticQrisDetailRequest struct {
	PaymentID  string
	MerchantID string
}
type StaticQrisListResponse struct {
	UUID                string     `json:"uuid" db:"uuid"`
	ReferenceID         string     `json:"referenceId" db:"reference_id"`
	MerchantID          string     `json:"merchantId" db:"merchant_id"`
	QrContent           string     `json:"qrContent" db:"qr_content"`
	QrUrl               string     `json:"qrUrl" db:"qr_url"`
	StoreID             string     `json:"storeId" db:"store_id"`
	QrImage             *string    `json:"qrImage,omitempty" db:"qr_image"`
	Status              string     `json:"status" db:"status"`
	CreatedAt           time.Time  `json:"createdAt" db:"created_at"`
	ExpiredAt           *time.Time `json:"expiredAt,omitempty" db:"expired_at"`
	StatementDescriptor *string    `json:"statementDescriptor,omitempty" db:"statement_descriptor"`
	MerchantName        string     `json:"merchantName" db:"merchant_name"`
}
type StaticQrisDetailResponse struct {
	UUID                string             `json:"uuid" db:"uuid"`
	ReferenceID         string             `json:"referenceId" db:"reference_id"`
	QrContent           string             `json:"qrContent" db:"qr_content"`
	QrUrl               string             `json:"qrUrl" db:"qr_url"`
	QrImage             *string            `json:"qrImage,omitempty" db:"qr_image"`
	Status              string             `json:"status" db:"status"`
	CreatedAt           time.Time          `json:"createdAt" db:"created_at"`
	ExpiredAt           *time.Time         `json:"expiredAt,omitempty" db:"expired_at"`
	TotalPayments       int                `json:"totalPayments" db:"total_payments"`
	TotalAmountValue    string             `json:"-" db:"total_amount"`
	StatementDescriptor *string            `json:"statementDescriptor,omitempty" db:"statement_descriptor"`
	TotalAmount         commonModel.Amount `json:"totalAmount"`
	MerchantName        string             `json:"merchantName" db:"merchant_name"`
}

// StaticQrisTransactionItem represents individual transaction in static QRIS
type StaticQrisTransactionItem struct {
	UUID            string     `json:"uuid" db:"uuid"`
	ReferenceID     string     `json:"referenceId" db:"reference_id"`
	AmountValue     string     `json:"-" db:"amount_value"`
	AmountCurrency  string     `json:"-" db:"amount_currency"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"createdAt" db:"created_at"`
	PaymentDate     *time.Time `json:"paymentDate,omitempty" db:"payment_date"`
	ProcessorRefID  string     `json:"processorReferenceId" db:"processor_reference_id"`
	BankReferenceID string     `json:"bankReferenceId" db:"bank_reference_id"`

	Amount commonModel.Amount `json:"amount"`
}

// StaticQrisTransactionFilterRequest represents the request for filtering transactions within a static QRIS
type StaticQrisTransactionFilterRequest struct {
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

// Default values for filtering
const (
	DefaultStaticQrisPage              = 1
	DefaultStaticQrisPerPage           = 12
	DefaultStaticQrisSort              = "DESC"
	DefaultStaticQrisSortBy            = "createdAt"
	DefaultStaticQrisTransactionSortBy = "paymentDate"
)

// Validate validates the StaticQrisFilterRequest
func (r *StaticQrisFilterRequest) Validate() {
	if r.Page < 1 {
		r.Page = DefaultStaticQrisPage
	}
	if r.PerPage < 1 {
		r.PerPage = DefaultStaticQrisPerPage
	}
	if r.Sort == "" {
		r.Sort = DefaultStaticQrisSort
	}
	if r.SortBy == "" {
		r.SortBy = DefaultStaticQrisSortBy
	}
}

// Validate validates the StaticQrisTransactionFilterRequest
func (r *StaticQrisTransactionFilterRequest) Validate() {
	if r.Page < 1 {
		r.Page = DefaultStaticQrisPage
	}
	if r.PerPage < 1 {
		r.PerPage = DefaultStaticQrisPerPage
	}
	if r.Sort == "" {
		r.Sort = DefaultStaticQrisSort
	}
	if r.SortBy == "" {
		r.SortBy = DefaultStaticQrisTransactionSortBy
	}
}

// StaticQrisUpdateStatusRequest represents the request for updating static QRIS status
type StaticQrisUpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE"`
}
