package paymentModel

import (
	"time"
)

type PaymentHistoryItem struct {
	UUID               string     `db:"uuid" json:"uuid"`
	ReferenceID        string     `db:"reference_id" json:"referenceId"`
	MerchantID         string     `db:"merchant_id" json:"merchantId"`
	RecurringID        *string    `db:"recurring_contract_id" json:"recurringId"`
	Method             string     `db:"payment_method" json:"paymentMethod"`
	MethodType         string     `db:"payment_method_type" json:"paymentMethodType"` // this property will support the frontend to get the subType like open_dynamic, dynamic, etc
	Channel            string     `db:"channel" json:"channel"`
	ProcessorRefNumber *string    `db:"processor_reference_number" json:"processorReferenceNumber"`
	Amount             string     `db:"amount" json:"amount"`
	AmountCurrency     string     `db:"amount_currency" json:"amountCurrency"`
	AmountPaid         *string    `db:"amount_paid" json:"amountPaid"`
	AmountPaidCurrency *string    `db:"amount_paid_currency" json:"amountPaidCurrency"`
	Status             string     `db:"status" json:"status"`
	CreatedAt          time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt          *time.Time `db:"updated_at" json:"updatedAt"` // latest settled payment information
	// Account transaction createdAt
	TransactionCreatedAt *time.Time `db:"transaction_created_at" json:"transactionCreatedAt"`
	HasSplitRouting      bool       `db:"has_split_routing" json:"hasSplitRouting"`
	HasInvestigation     bool       `db:"has_investigation" json:"hasInvestigation"`
	SettlementModel      string     `db:"settlement_model" json:"settlementModel"`
	PaidAt               *time.Time `json:"paidAt" db:"paid_at"`
	PaymentLink          string     `db:"payment_url" json:"paymentLink,omitempty"`

	// Refund summary information
	RefundInfo *RefundInfo `json:"refundInfo,omitempty" db:"-"`

	// Support Payment History Download
	ExpiredAt          *time.Time `json:"-" db:"expired_at"`
	CustomerId         string     `json:"-" db:"customer_id"`
	VirtualAccountNo   string     `json:"-" db:"virtual_account_no"`
	VirtualAccountName string     `json:"-" db:"virtual_account_name"`
	QrisMerchantName   string     `json:"-" db:"qris_merchant_name"`
	QrisURL            string     `json:"-" db:"qris_url"`
	CardType           string     `json:"-" db:"credit_card_type"`
	CardIssuerBank     string     `json:"-" db:"credit_card_issuer_bank"`
	CardNumber         string     `json:"-" db:"credit_card_number"`
	CardExpiry         string     `json:"-" db:"credit_card_expiry"`
	// Refund information
	RefundDate          *time.Time `json:"-" db:"refund_date"`
	RefundAmount        *string    `json:"-" db:"refund_amount"`
	RefundStatus        *string    `json:"-" db:"refund_status"`
	TotalRefundedAmount *string    `json:"-" db:"total_refunded_amount"`
	RefundCount         *int       `json:"-" db:"refund_count"`
	// Support Payment History Download - MID Info
	MID         string `json:"-" db:"mid"`
	MIDType     string `json:"-" db:"mid_type"`
	MIDAcquirer string `json:"-" db:"mid_acquirer"`
}

type RefundInfo struct {
	Type                string `json:"type"`                          // NONE, PARTIAL, FULL
	TotalRefundedAmount string `json:"totalRefundedAmount,omitempty"` // Sum of all successful refunds
	RefundCount         int    `json:"refundCount,omitempty"`         // Number of refunds
}

type PaymentInsightQueryResult struct {
	Total       int     `db:"total_payment"`
	TotalAmount float64 `db:"total_amount"`
}
