package cardFundedPayoutModel

import (
	"time"

	common "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/fee"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type CreateSavedCardResponse struct {
	ReferenceID string `json:"referenceId"`
	PaymentUrl  string `json:"paymentUrl"`
}

type GetSavedCardResponse struct {
	ID             string `json:"id" db:"uuid"`
	CardName       string `json:"cardName" db:"card_name"`
	CardOrigin     string `json:"cardOrigin" db:"card_origin"` // LOCAL or FOREIGN
	PaymentChannel string `json:"paymentChannel" db:"payment_channel"`
	IssuingBank    string `json:"issuingBank" db:"issuing_bank"`
	Last4          string `json:"last4" db:"last_4_digits"`
	ExpiryMonth    string `json:"expiryMonth" db:"expiry_month"`
	ExpiryYear     string `json:"expiryYear" db:"expiry_year"`
	// For internal use
	MerchantName string `json:"-" db:"merchant_name"`
	CardToken    string `json:"-" db:"card_token"`
}

type PayoutActionResponse struct {
	ID                string               `json:"id"`
	VendorID          string               `json:"vendorId"`
	VendorName        string               `json:"vendorName"`
	ReferenceID       string               `json:"referenceId"`
	BankCode          string               `json:"bankCode,omitempty"`
	BankName          string               `json:"bankName,omitempty"`
	AccountNumber     string               `json:"accountNumber,omitempty"`
	AccountName       string               `json:"accountName,omitempty"`
	FeeAmount         float64              `json:"feeAmount"`
	Amount            common.AmountRequest `json:"amount"`
	Remarks           string               `json:"remarks"`
	SettlementMethod  string               `json:"settlementMethod"`
	CardID            string               `json:"cardId"`
	CardName          string               `json:"cardName"`
	AuthenticationUrl *string              `json:"authenticationUrl,omitempty"`
	RejectReason      *string              `json:"rejectReason,omitempty"`
	CreatedAt         *time.Time           `json:"createdAt,omitempty"`
	ApprovedAt        *time.Time           `json:"approvedAt,omitempty"`
	RejectedAt        *time.Time           `json:"rejectedAt,omitempty"`
	MerchantID        string               `json:"merchantId,omitempty"`
	Status            string               `json:"status,omitempty"`
	BankReferenceNo   string               `json:"bankReferenceNo,omitempty"`
	ReconReferenceNo  string               `json:"reconReferenceNo,omitempty"`
	ReasonType        *string              `json:"reasonType,omitempty"`
	ReasonDescription *string              `json:"reasonDescription,omitempty"`
}

type TransactionConfigResponse struct {
	FeeDetail feeModel.FeeResponseder `json:"feeDetail"`
}

type GetPayoutListResponse struct {
	UUID              string    `json:"uuid" db:"uuid"`
	CreatedAt         time.Time `json:"createdAt" db:"created_at"`
	ReferenceID       string    `json:"referenceId" db:"reference_id"`
	Amount            string    `json:"amount" db:"amount"`
	Fee               string    `json:"fee" db:"fee"`
	TotalAmount       string    `json:"totalAmount" db:"total_amount"`
	TransactionStatus string    `json:"transactionStatus" db:"transaction_status"` // PROCESSING/SUCCESS/FAILED
	ApprovalStatus    string    `json:"approvalStatus" db:"status"`                // WAITING/APPROVED/REJECTED
	VendorID          string    `json:"vendorId" db:"-"`
	VendorName        string    `json:"vendorName" db:"-"`
	Remarks           string    `json:"remarks" db:"remark"`
	BankName          string    `json:"bankName" db:"beneficiary_bank_name"`
	AccountNumber     string    `json:"accountNumber" db:"beneficiary_account_no"`
	AccountName       string    `json:"accountName" db:"beneficiary_account_name"`
	Card              CardInfo  `json:"card" db:"-"`

	// For internal use
	Metadata    types.NullJSONText         `json:"-" db:"metadata"`
	MetadataObj disbursementModel.Metadata `json:"-" db:"-"`
}

func (r *GetPayoutListResponse) Hydrate() {
	if r.Metadata.Valid {
		_ = r.Metadata.Unmarshal(&r.MetadataObj)
	}

	if r.MetadataObj.CardFundedDetail != nil {
		r.VendorID = r.MetadataObj.CardFundedDetail.VendorID
		r.VendorName = r.MetadataObj.CardFundedDetail.VendorName

		if r.MetadataObj.CardFundedDetail.Card != nil {
			card := r.MetadataObj.CardFundedDetail.Card
			expiry := card.ExpiryMonth + "/" + card.ExpiryYear
			if len(card.ExpiryYear) >= 2 {
				expiry = card.ExpiryMonth + "/" + card.ExpiryYear[len(card.ExpiryYear)-2:]
			}
			r.Card = CardInfo{
				LastFour: card.Last4Digits,
				Brand:    card.CardName,
				Channel:  card.PaymentChannel,
				Name:     card.CardName,
				Issuer:   card.IssuingBank,
				Expiry:   expiry,
			}
		}
	}
}

type CardInfo struct {
	LastFour string `json:"lastFour"`
	Brand    string `json:"brand"`  // cardName
	Channel  string `json:"type"`   // DEBIT | CREDIT
	Name     string `json:"name"`   // cardName
	Issuer   string `json:"issuer"` // issuingBank
	Expiry   string `json:"expiry"` // MM/YY format
}

// GetPayoutDetailResponse is the response for payout detail
type GetPayoutDetailResponse struct {
	UUID              string              `json:"uuid" db:"uuid"`
	CreatedAt         time.Time           `json:"createdAt" db:"created_at"`
	ReferenceID       string              `json:"referenceId" db:"reference_id"`
	Amount            string              `json:"amount" db:"amount"`
	Fee               string              `json:"fee" db:"fee"`
	TotalAmount       string              `json:"totalAmount" db:"total_amount"`
	TransactionStatus string              `json:"transactionStatus" db:"transaction_status"` // PROCESSING/SUCCESS/FAILED
	ApprovalStatus    string              `json:"approvalStatus" db:"status"`                // WAITING/APPROVED/REJECTED
	VendorID          string              `json:"vendorId" db:"-"`
	VendorName        string              `json:"vendorName" db:"-"`
	Remarks           string              `json:"remarks" db:"remark"`
	BankName          string              `json:"bankName" db:"beneficiary_bank_name"`
	AccountNumber     string              `json:"accountNumber" db:"beneficiary_account_no"`
	AccountName       string              `json:"accountName" db:"beneficiary_account_name"`
	Card              CardInfo            `json:"card" db:"-"`
	ChargeIDs         []string            `json:"chargeIds" db:"-"`
	ApprovalDate      *time.Time          `json:"approvalDate,omitempty" db:"approved_at"`
	ApprovedBy        *string             `json:"approvedBy,omitempty" db:"approved_by"`
	CurrentStatus     string              `json:"currentStatus" db:"-"`
	StatusHistory     []StatusHistoryItem `json:"statusHistory" db:"-"`

	// For internal use
	Metadata    types.NullJSONText         `json:"-" db:"metadata"`
	MetadataObj disbursementModel.Metadata `json:"-" db:"-"`
	MerchantID  string                     `json:"-" db:"merchant_id"`
}

// StatusHistoryItem represents a single status history entry
type StatusHistoryItem struct {
	Status      string     `json:"status"`
	Label       string     `json:"label"`
	Description string     `json:"description,omitempty"`
	Timestamp   *time.Time `json:"timestamp,omitempty"`
}

// ExportPayoutListResponse is the response for export payout list
type ExportPayoutListResponse struct {
	Url string `json:"url"`
}

// GetPayoutInsightsResponse is the response for card-funded payout insights.
// Returns aggregated totals for payouts with WAITING approval status.
type GetPayoutInsightsResponse struct {
	TotalAmount      common.Amount `json:"totalAmount"`
	TotalTransaction int           `json:"totalTransaction"`
}

// GetPayoutInsightsDTO is the intermediate DTO from the database query.
type GetPayoutInsightsDTO struct {
	Count int     `db:"count"`
	Sum   float64 `db:"sum"`
}

// Hydrate populates derived fields from metadata
func (r *GetPayoutDetailResponse) Hydrate() {
	if r.Metadata.Valid {
		_ = r.Metadata.Unmarshal(&r.MetadataObj)
	}

	if r.MetadataObj.CardFundedDetail != nil {
		r.VendorID = r.MetadataObj.CardFundedDetail.VendorID
		r.VendorName = r.MetadataObj.CardFundedDetail.VendorName

		if r.MetadataObj.CardFundedDetail.Card != nil {
			card := r.MetadataObj.CardFundedDetail.Card
			expiry := card.ExpiryMonth + "/" + card.ExpiryYear
			if len(card.ExpiryYear) >= 2 {
				expiry = card.ExpiryMonth + "/" + card.ExpiryYear[len(card.ExpiryYear)-2:]
			}
			r.Card = CardInfo{
				LastFour: card.Last4Digits,
				Brand:    card.CardName,
				Channel:  card.PaymentChannel,
				Name:     card.CardName,
				Issuer:   card.IssuingBank,
				Expiry:   expiry,
			}
		}
	}

	// Set current status based on transaction status or approval
	r.CurrentStatus = r.TransactionStatus
	if r.CurrentStatus == "" {
		r.CurrentStatus = r.ApprovalStatus
	}
}

// GetReceiptRequest is the request for getting payout receipt
type GetReceiptRequest struct {
	PayoutID   string `json:"-"`
	MerchantID string `json:"-"`
}

// GetReceiptResponse is the response for payout receipt
type GetReceiptResponse struct {
	ReceiptURL string `json:"receiptUrl"`
}

// ReceiptData is the data structure for receipt template
type ReceiptData struct {
	// Transaction Detail
	CreatedAt   string
	ReferenceID string
	PayoutID    string
	Amount      string
	Fee         string
	TotalAmount string

	// Recipient Detail
	VendorName  string
	Remarks     string
	BankName    string
	AccountNo   string
	AccountName string

	// Images
	ImageHeader     string
	ImageBackground string
}

type GetPayoutTransactionListResponse struct {
	ID                string          `json:"id" db:"id"`
	TrxID             string          `json:"trxId" db:"trx_id"`
	ClientReferenceID string          `json:"clientReferenceId" db:"client_reference_id"`
	VendorID          string          `json:"vendorId" db:"vendor_id"`
	VendorName        string          `json:"vendorName" db:"vendor_name"`
	BankCode          string          `json:"bankCode" db:"bank_code"`
	BankName          string          `json:"bankName" db:"bank_name"`
	AccountNumber     string          `json:"accountNumber" db:"account_number"`
	AccountName       string          `json:"accountName" db:"account_name"`
	Remarks           string          `json:"remarks" db:"remarks"`
	TrxAmount         decimal.Decimal `json:"trxAmount" db:"trx_amount"`
	TrxStatus         string          `json:"trxStatus" db:"trx_status"`
	TrxReasonType     *string         `json:"trxReasonType" db:"trx_reason_type"`
	TrxReasonDesc     *string         `json:"trxReasonDesc" db:"trx_reason_desc"`
	CreatedAt         time.Time       `json:"createdAt" db:"created_at"`
	ApprovedAt        time.Time       `json:"approvedAt" db:"approved_at"`
	ScheduledAt       *time.Time      `json:"scheduledAt" db:"scheduled_at"`
	TrxCreatedAt      *time.Time      `json:"trxCreatedAt" db:"trx_created_at"`
	TrxUpdatedAt      *time.Time      `json:"trxUpdatedAt" db:"trx_updated_at"`
	MerchantID        string          `json:"merchantId" db:"merchant_id"`
	InitAmount        decimal.Decimal `json:"initAmount" db:"init_amount"`
	InitFee           decimal.Decimal `json:"initFee" db:"init_fee"`
	InitTotalAmount   decimal.Decimal `json:"initTotalAmount" db:"init_total_amount"`
	ExecAmount        decimal.Decimal `json:"execAmount" db:"exec_amount"`
	ExecFee           decimal.Decimal `json:"execFee" db:"exec_fee"`
	ExecTotalAmount   decimal.Decimal `json:"execTotalAmount" db:"exec_total_amount"`
}
