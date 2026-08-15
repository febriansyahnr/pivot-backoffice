package disbursementModel

import (
	"io"
	"time"

	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/proto/messages/disbursement"

	"github.com/shopspring/decimal"
)

type CreateSingleRequest struct {
	ReferenceID            string          `json:"referenceId" validate:"required" example:"client-ref-from-client"`
	BeneficiaryBankCode    string          `json:"beneficiaryBankCode" validate:"required" example:"008"`
	BeneficiaryBankName    string          `json:"beneficiaryBankName" validate:"required" example:"Bank 008"`
	BeneficiaryAccountNo   string          `json:"beneficiaryAccountNo" validate:"required" example:"8000800808"`
	BeneficiaryAccountName string          `json:"beneficiaryAccountName" validate:"required" example:"Julio Caesar"`
	Amount                 decimal.Decimal `json:"amount" validate:"required" example:"100000"`
	Remark                 string          `json:"remark" validate:"max=40" example:"ASPI testing"`
	PurposeID              string          `json:"purposeId" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	InquiryID              string          `json:"inquiryId" example:"uuid-uuid-uuid-uuid"`

	BulkID       *string `json:"-"`
	MerchantID   string  `json:"-"`
	MerchantName string  `json:"-"`
	CreatedBy    *string `json:"-"`
	CreatedFrom  string  `json:"-"`
}

type ExportDisbursementFilterRequest struct {
	StartCreatedAt    string `json:"startCreatedAt"`
	EndCreatedAt      string `json:"endCreatedAt"`
	Status            string `json:"status"`
	TransactionStatus string `json:"transactionStatus"`
	Type              string `json:"type"`
	Keyword           string `json:"keyword"`
	Sort              string `json:"sort"`
	SortBy            string `json:"sortBy"`
}

type GetDisbursementFilterRequest struct {
	MerchantID        string     `json:"-"`
	UUID              string     `json:"uuid"`
	StartCreatedAt    *time.Time `json:"startCreatedAt"`
	EndCreatedAt      *time.Time `json:"endCreatedAt"`
	BulkID            string     `json:"bulkId"`
	Status            string     `json:"status"`
	Type              string     `json:"type"`
	TransactionStatus string     `json:"transactionStatus"`
	Keyword           string     `json:"keyword"`
	Sort              string     `json:"sort"`
	SortBy            string     `json:"sortBy"`

	IsXbPayout bool   `json:"-"`
	ReasonType string `json:"-"`
}

func (r *GetDisbursementFilterRequest) Validate() error {
	if r.Type != "" {
		if err := ValidateDisbursementType(r.Type); err != nil {
			return err
		}
	}

	if r.Status != "" {
		if err := ValidateApprovalStatus(r.Status); err != nil {
			return err
		}
	}

	if r.TransactionStatus != "" {
		if err := ValidateTransactionStatus(r.TransactionStatus); err != nil {
			return err
		}
	}

	if r.SortBy != "" {
		if err := ValidateSortColumn(r.SortBy); err != nil {
			return err
		}
	}

	return nil
}

type ApprovalActionsRequest struct {
	BulkID  string                `json:"bulkId" validate:"omitempty,uuid"`
	Approve []ApproveActionObject `json:"approve" validate:"omitempty,dive,required"`
	Reject  []RejectActionObject  `json:"reject" validate:"omitempty,dive,required"`

	UserID     string `json:"-" validate:"-"`
	MerchantID string `json:"-" validate:"-"`
}

type ApproveRequest struct {
	ApproveAction []ApproveActionObject `json:"-"`
	MerchantID    string                `json:"-"`
	ApprovedBy    string                `json:"-"`
	BulkID        string                `json:"bulkId"`
	CreatedFrom   string                `json:"createdFrom"`
	TotalAmount   float64               `json:"-"` // Total Amount Without Fee
	IsCompleted   bool                  `json:"-"` // Status update on approval process
}

type RejectRequest struct {
	RejectAction []RejectActionObject `json:"-"`
	MerchantID   string               `json:"-"`
	RejectedBy   string               `json:"-"`
	BulkID       string               `json:"bulkId"`
}

type ValidateBalanceRequest struct {
	DisbursementIDs []string `json:"-"`
	MerchantID      string   `json:"-"`
}

type RetrySingleRequest struct {
	DisbursementID string `json:"id" validate:"required" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID     string `json:"-"`
}

type RetryBulkRequest struct {
	BulkDisbursementID string `json:"id" validate:"required" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	MerchantID         string `json:"-"`
}

type CreateDisbursementFromOpenApiRequest struct {
	Payouts []PayoutObjectForCreate `json:"payouts" validate:"required,min=1"`
}

type DisbursementDataCallbackRequest struct {
	MerchantID string `json:"merchantId"`

	// For Bulk Payout and Existing Single Payout
	UUID          string              `json:"uuid,omitempty"`
	PayoutResults *PayoutResultObject `json:"payoutResults,omitempty"`
	Status        string              `json:"status,omitempty"`

	// For Single Payout
	Payout *PayoutCallbackSingleObject `json:"payout,omitempty"`
}

type RetryTransaction struct {
	DisbursementID string `json:"-"`
	MerchantID     string `json:"merchantId" validate:"required,uuid"`
	ForceFailed    bool   `json:"forceFailed"` // Force NOT_FOUND status to FAILED in SNAP Core (default: false)
	// will force retry at processor, but still check transfer process status
	ForceRetry bool `json:"forceRetry"` // Force retry even if already succeeded (default: false)
	// will bypass processor check, only use when you know what you're doing
	BypassProcessorCheck bool `json:"bypassProcessorCheck"` // Bypass processor check (default: false)
}

type CancelPayoutRequest struct {
	MerchantID  string   `json:"-"`
	BatchBulkID []string `json:"batchBulkId"`
	BatchID     []string `json:"batchId"`
}

type InquiryTransaction struct {
	DisbursementID string `json:"-"`
	MerchantID     string `json:"merchantId" validate:"required,uuid"`
}

type ReversalTransactionReq struct {
	DisbursementId string `json:"-" validate:"required,uuid"`
	MerchantId     string `json:"merchantId" validate:"required,uuid"`
	CreatedBy      string `json:"createdBy" validate:"required,max=255"`
	Reason         string `json:"reason" validate:"required,max=255"`
}

type CRMSinglePayoutStatusRequest struct {
	ReferenceID string `json:"referenceId" validate:"required"`
}

type CRMBatchPayoutStatusRequest struct {
	ReferenceIDs []string `json:"referenceIds" validate:"required,min=1"`
}

type BulkPreviewRequest struct {
	MerchantId string
	File       io.Reader
}

type BulkCreateRequest struct {
	MerchantId string    `json:"merchantId"`
	CreatedBy  string    `json:"createdBy"`
	File       io.Reader `json:"-"`
}

func TransformArrayCreateSingleRequestToProtobufType(data []CreateSingleRequest) []*pb.CreateSingleRequest {

	result := make([]*pb.CreateSingleRequest, len(data))

	for i, d := range data {
		result[i] = &pb.CreateSingleRequest{
			ReferenceId:            d.ReferenceID,
			BeneficiaryBankCode:    d.BeneficiaryBankCode,
			BeneficiaryBankName:    d.BeneficiaryBankName,
			BeneficiaryAccountNo:   d.BeneficiaryAccountNo,
			BeneficiaryAccountName: d.BeneficiaryAccountName,
			Amount:                 d.Amount.String(),
			Remark:                 d.Remark,
			PurposeId:              d.PurposeID,
			InquiryId:              d.InquiryID,
		}
	}
	return result
}

type ChangeDisbursementTransactionStatusRequest struct {
	DisbursementIDS   []string `json:"disbursementIds" validate:"required,min=1,dive,uuid"`
	Status            string   `json:"status" validate:"required,oneof=SUCCESS PENDING FAILED"`
	ReasonType        *string  `json:"reasonType" validate:"required_if=Status FAILED,omitempty,oneof=OTHER INSUFFICIENT_ESCROW_FUND INVALID_ACCOUNT BLOCKED_BY_HARSYA REVERSAL BLOCKED_BY_BANK"`
	ReasonDescription *string  `json:"reasonDescription,omitempty"`
	ReferenceNumber   string   `json:"referenceNumber" validate:"omitempty,max=60"`
}

type CheckDisbursementTransactionStatusRequest struct {
	DisbursementIDs []string `json:"disbursementIds" validate:"required,min=1,dive,uuid"`
}

type BeneficiaryPayoutLimitAlertRequest struct {
	TotalAmount              float64
	NumberOfTransaction      int64
	AmountThreshold          float64
	CountThreshold           int64
	BeneficiaryAccountNumber string
	BeneficiaryAccountName   string
	BeneficiaryBankCode      string
	MerchantID               string
}

type PayoutTransactionAlertRequest struct {
	DisbursementID string                           `json:"disbursementId"`
	BankRefNo      string                           `json:"bankRefNo"`
	BankProcessor  string                           `json:"bankProcessor"`
	TransferType   string                           `json:"transferType"`
	History        []*PayoutTransactionAlertHistory `json:"history,omitempty"`
}

type PayoutTransactionAlertHistory struct {
	Acquirer string `json:"acquirer"`
	Order    int    `json:"order"`
	Status   string `json:"status"`
	Time     string `json:"time"`
	Action   string `json:"action"`
	Reason   string `json:"reason,omitempty"`
}

type GetDisbursementReceiptRequest struct {
	DisbursementID string
	ReferenceID    string
	MerchantID     string
}

type GetDisbursementReceiptCRMRequest struct {
	ReferenceID string `json:"referenceId" validate:"required"`
	MerchantID  string `json:"merchantId" validate:"required,uuid"`
}

type GetXbPayoutDashboardInsightRequest struct {
	StartDate  time.Time
	EndDate    time.Time
	MerchantId string
}
