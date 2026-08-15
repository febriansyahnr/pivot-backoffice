package disbursementModel

import (
	"database/sql"
	"encoding/json"
	"slices"
	"strings"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/statusHistory"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/xb"
)

type Disbursement struct {
	UUID                   string             `json:"uuid" db:"uuid" example:"b7b95825-6ec3-486a-a3cf-62d4ed815757"`
	ReferenceID            string             `json:"referenceId" db:"reference_id" example:"reference-id-from-client"`
	MerchantID             string             `json:"merchantId" db:"merchant_id" example:"50f95963-373a-4a6e-a7c9-79b6cf891f9c"`
	BulkID                 *string            `json:"bulkId" db:"bulk_id" example:"ff35c906-0144-4a52-8621-4991dd00f945"`
	PurposeID              *string            `json:"purposeId" db:"purpose_id" example:"bad7bd6f-3459-4e28-97d5-e200469d5cc6"`
	Type                   string             `json:"type" db:"type"` // Auto-generated from metadata. Do not set this value during insert, as it will be overwritten.
	SenderName             string             `json:"senderName" db:"sender_name" example:"Google Corp"`
	AccountInquiryID       *string            `json:"-" db:"account_inquiry_id" example:"ff35c906-0144-4a52-8621-4991dd009999"`
	BeneficiaryBankCode    string             `json:"beneficiaryBankCode" db:"beneficiary_bank_code" example:"008"`
	BeneficiaryBankName    *string            `json:"beneficiaryBankName" db:"beneficiary_bank_name" example:"Bank Permata"`
	BeneficiaryAccountNo   string             `json:"beneficiaryAccountNo" db:"beneficiary_account_no" example:"8000800808"`
	BeneficiaryAccountName string             `json:"beneficiaryAccountName" db:"beneficiary_account_name" example:"Yories Yolanda"`
	ProcessorReferenceID   *string            `json:"processorReferenceId" db:"processor_reference_id" example:"dce9c63e-8a1c-4d8f-b18d-33de6f772104"`
	ProcessorReferenceName *string            `json:"processorReferenceName" db:"processor_reference_name" example:"John Doe"`
	BankReferenceNo        *string            `json:"bankReferenceNo" db:"bank_reference_no" example:"WErft7089990"`
	Currency               string             `json:"currency" db:"currency" example:"IDR"`
	Amount                 decimal.Decimal    `json:"amount" db:"amount" example:"1000000.00"`
	Fee                    *decimal.Decimal   `json:"fee" db:"fee" example:"4000.00"`
	TotalAmount            decimal.Decimal    `json:"totalAmount" db:"total_amount" example:"1004000.00"`
	Status                 string             `json:"status" db:"status"  example:"APPROVED"`
	ReasonType             *string            `json:"reasonType" db:"reason_type" example:"OTHER"`
	ReasonDescription      *string            `json:"reasonDescription" db:"reason_description" example:"User input"`
	Remark                 *string            `json:"remark" db:"remark" example:"Disburse to Yories"`
	Metadata               types.NullJSONText `json:"-" db:"metadata"`
	CreatedFrom            *string            `json:"createdFrom" db:"created_from" example:"MERCHANT_PORTAL | OPEN_API"`
	CreatedBy              *string            `json:"createdBy" db:"created_by" example:"John Doe"`
	ApprovedBy             *string            `json:"approvedBy" db:"approved_by" example:"Lorem Ipsum"`
	ApprovedAt             *time.Time         `json:"approvedAt" db:"approved_at" example:"2021-01-01T00:00:00Z"`
	CreatedAt              time.Time          `json:"createdAt" db:"created_at" example:"2021-01-01T00:00:00Z"`
	UpdatedAt              time.Time          `json:"updatedAt" db:"updated_at" example:"2021-01-01T00:00:00Z"`
	DeletedAt              *time.Time         `json:"deletedAt" db:"deleted_at" example:"2021-01-01T00:00:00Z"`
	// Internal Data
	MetadataObj Metadata `json:"metadata" db:"-"`
}

func (d *Disbursement) GetCardFundedPayoutDetail() *CardFundedDetailMetadata {
	if d.MetadataObj.CardFundedDetail == nil {
		return &CardFundedDetailMetadata{}
	}
	return d.MetadataObj.CardFundedDetail
}

type Metadata struct {
	FeeDetail        feeModel.FeeMetadataObject       `json:"feeDetail"`
	XbDetail         *xbModel.XbPayoutMetadata        `json:"xbDetail,omitempty"`
	FeeOnBehalf      *feeModel.TrxFeeOnBehalfMetadata `json:"feeOnBehalf,omitempty"`
	OnBehalf         *merchantModel.OnBehalfObject    `json:"onBehalf,omitempty"`
	CardFundedDetail *CardFundedDetailMetadata        `json:"cardFundedDetail,omitempty"`
}

type CardFundedDetailMetadata struct {
	VendorID         string                        `json:"vendorId"`
	VendorName       string                        `json:"vendorName"`
	Card             *CardFundedDetailMetadataCard `json:"card"`
	SettlementMethod string                        `json:"settlementMethod"`
}

type CardFundedDetailMetadataCard struct {
	ID             string `json:"id"`
	CardName       string `json:"cardName"`
	PaymentChannel string `json:"paymentChannel"`
	IssuingBank    string `json:"issuingBank"`
	Last4Digits    string `json:"last4Digits"`
	ExpiryMonth    string `json:"expiryMonth"`
	ExpiryYear     string `json:"expiryYear"`
}

type DisbursementBulk struct {
	UUID        string          `json:"uuid" db:"uuid"`
	MerchantID  string          `json:"merchantId" db:"merchant_id"`
	File        string          `json:"file" db:"file"`
	Status      string          `json:"status" db:"status"`
	Currency    string          `json:"currency" db:"currency"`
	TotalAmount decimal.Decimal `json:"totalAmount" db:"total_amount"`
	TotalFee    decimal.Decimal `json:"totalFee" db:"total_fee"`
	CreatedAt   time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time       `json:"updatedAt" db:"updated_at"`
	DeletedAt   *time.Time      `json:"deletedAt" db:"deleted_at"`
}

type SumAmountResponse struct {
	TotalAmount      float64 `json:"-" db:"sum_total_amount"`
	ParentFeeCharged float64 `json:"-" db:"sum_parent_fee_charged"`
}

type ActionTransactionSummary struct {
	Total       int     `db:"total"`
	TotalAmount float64 `db:"total_amount"`
}

type DisbursementWithTransaction struct {
	Disbursement
	TransactionStatus               *string `json:"transactionStatus" db:"transaction_status"`
	TransactionReasonType           *string `json:"transactionReasonType" db:"transaction_reason_type"`
	TransactionReasonDescription    *string `json:"transactionReasonDesc,omitempty" db:"transaction_reason_description"`
	TransactionProcessorReferenceID *string `json:"-" db:"transaction_processor_reference_id"`
	TransactionProcessorReference   *string `json:"-" db:"processor_reference"`

	// This field is solely used to support the creation of pending ledgers for the Payout Cut-off Time feature.
	CutOffTimeStatusOngoing bool `json:"-" db:"-"`

	// This field only fetch from get by detail
	StatusHistories []*statusHistoriesModel.StatusHistory `json:"-" db:"-"`
}

type TransactionConfig struct {
	MinAmount float64 `json:"minAmount" redis:"minAmount"`
	MaxAmount float64 `json:"maxAmount" redis:"maxAmount"`
}

func (t TransactionConfig) MarshalBinary() ([]byte, error) {
	return json.Marshal(t)
}

func (t *TransactionConfig) UnmarshalBinary(buf []byte) error {
	return json.Unmarshal(buf, t)
}

type DisbursementForReversal struct {
	Id             string         `db:"id"`
	Status         string         `db:"status"`
	ReasonType     sql.NullString `db:"reason_type"`
	MerchantId     string         `db:"merchant_id"`
	Currency       string         `db:"currency"`
	Amount         float64        `db:"amount"`
	FeeAmount      float64        `db:"fee_amount"`
	TotalAmount    float64        `db:"total_amount"`
	RawTransaction types.JSONText `db:"transaction"`
	RawFee         types.JSONText `db:"fee"`
	// Internal Data
	Transaction TransactionMetadataForReversal `db:"-"`
	Fee         TransactionMetadataForReversal `db:"-"`
}

func (d *DisbursementForReversal) IsFeeStatus(status string) bool {
	return d.Fee.Status == status
}

func (d *DisbursementForReversal) IsFeeDeductionType(types ...string) bool {
	return slices.Contains(types, d.Fee.Metadata.DeductionType)
}

type TransactionMetadataForReversal struct {
	Id       string                     `json:"id"`
	Amount   float64                    `json:"amount"`
	Status   string                     `json:"status"`
	Metadata feeModel.FeeMetadataObject `json:"metadata"`
}

type ReversalMetadataObject struct {
	TransactionId string `json:"transactionId"`
	FeeId         string `json:"feeId"`
}

type CutOffTimeStatusResponse struct {
	Status      string    `json:"status"`
	Time        string    `json:"time,omitempty"` // Using Config Timezone
	Banner      string    `json:"banner,omitempty"`
	ProcessedAt time.Time `json:"-"` // Using UTC
}

type BeneficiaryPayoutLimitRuleConfig struct {
	Velocity        int64   `json:"velocity"`
	Timeframe       string  `json:"timeframe"` // Do not read it for now, default = DAILY
	AmountThreshold float64 `json:"amountThreshold"`
}

type BeneficiaryPayoutLimitRuleLimit struct {
	BeneficiaryPayoutLimitRuleConfig

	Count     int     `json:"count" db:"count_payout" redis:"count"`
	Processed float64 `json:"processed" db:"processed" redis:"processed"`
}

type ApprovalValidation struct {
	AccountNo string  `json:"accountNo"`
	Amount    float64 `json:"amount"`
	Error     error   `json:"-"`
}

type ApprovalResultErr struct {
	BeneficiaryLimitExceeded []ApprovalValidation `json:"beneficiaryLimitExceeded"`
}

func (e *ApprovalResultErr) Error() string {
	return constant.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions
}

type MerchantIDForPayoutCallback struct {
	MerchantId       string `db:"merchant_id"`
	ParentMerchantId string `db:"parent_merchant_id"`
}

func ValidateDisbursementType(disbursementType string) error {
	disbursementType = strings.ToUpper(disbursementType)
	switch disbursementType {
	case constant.DisbursementTypeSingle, constant.DisbursementTypeBulk:
		return nil
	default:
		return constant.ErrInvalidDisbursementType
	}
}

func ValidateApprovalStatus(approvalStatus string) error {
	approvalStatus = strings.ToUpper(approvalStatus)
	switch approvalStatus {
	case constant.DisbursementStatusPending, constant.DisbursementStatusWaiting, constant.DisbursementStatusApproved, constant.DisbursementStatusRejected:
		return nil
	default:
		return constant.ErrInvalidDisbursementApprovalStatus
	}
}

func ValidateTransactionStatus(status string) error {
	status = strings.ToUpper(status)
	switch status {
	case constant.StatusPending, constant.StatusSuccess, constant.StatusFailed, constant.SettlementStatusCancelled:
		return nil
	default:
		return constant.ErrInvalidDisbursementPaymentStatus
	}
}

func ValidateSortColumn(column string) error {
	switch column {
	case "updatedAt", "createdAt":
		return nil
	default:
		return constant.ErrInvalidDisbursementListSortColumn
	}
}

type ReconfirmXBRequest struct {
	PayoutId     string    `json:"payoutId"`
	XBStatus     string    `json:"xbStatus"`
	ExtendedTime time.Time `json:"extendedTime"`
}

type XbPayoutDashboardInsights struct {
	WaitingForConfirmCount    uint                                 `db:"waiting_for_confirm_count" json:"waitingForConfirmCount"`
	InformationRequestedCount uint                                 `db:"information_requested_count" json:"informationRequestedCount"`
	PendingCount              uint                                 `db:"pending_count" json:"pendingCount"`
	SuccessCount              uint                                 `db:"success_count" json:"successCount"`
	SuccessTotal              json.Number                          `db:"success_total" json:"successTotal"`
	RawTopCountriesByVolume   types.NullJSONText                   `db:"top_countries_by_volume" json:"-"`
	TopCountriesByVolume      []XbPayoutTransactionVolumeByCountry `db:"-" json:"topCountriesByVolume"`
}

type XbPayoutTransactionVolumeByCountry struct {
	Country    string      `json:"country"`
	Volume     json.Number `json:"volume"` // Value in IDR currency
	Percentage json.Number `json:"percentage"`
}
