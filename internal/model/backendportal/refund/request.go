package refundModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
)

type CreateRefundRequest struct {
	ClientReferenceID   string               `json:"clientReferenceId" validate:"required"`
	PaymentSessionID    string               `json:"paymentSessionId" validate:"required"`
	ChargeID            string               `json:"chargeId"`
	IsFullAmount        bool                 `json:"isFullAmount"`
	Amount              *commonModel.Amount  `json:"amount" validate:"required_if=IsFullAmount false,omitempty"`
	Reason              string               `json:"reason" validate:"required,oneof=SUSPECT_FRAUDULENT DUPLICATE REQUESTED_BY_CUSTOMER CANCELLATION OTHERS"`
	Description         string               `json:"description" validate:"max=50"`
	Method              string               `json:"method" validate:"required,oneof=AUTO TRANSFER_ONLY"`
	TransferDestination *TransferDestination `json:"transferDestination,omitempty" validate:"required_if=Method TRANSFER_ONLY"`
	Metadata            interface{}          `json:"metadata,omitempty" validate:"omitempty,dive"`

	MerchantID   string `json:"-"`
	IsCRMRequest bool   `json:"-"`
}

func NewCreatRefundRequest() *CreateRefundRequest {
	return &CreateRefundRequest{
		IsFullAmount: false,
		Method:       constant.RefundMethodAuto,
		IsCRMRequest: false,
	}
}

type CreateRefundThroughCRMRequest struct {
	CreateRefundRequest

	MerchantID string `json:"merchantId" validate:"required"`
	Method     string `json:"method" validate:"required,oneof=AUTO"`
}

func NewCreatRefundThroughCRMRequest() *CreateRefundThroughCRMRequest {
	return &CreateRefundThroughCRMRequest{
		CreateRefundRequest: CreateRefundRequest{
			IsFullAmount: false,
			Method:       constant.RefundMethodAuto,
			IsCRMRequest: true,
		},
	}
}

type RefundProcessRequest struct {
	RefundID string `json:"refundId"`

	PaymentMethodChannelType      string  `json:"paymentMethodChannelType"`
	PaymentMethodType             string  `json:"-"`
	PaymentProcessorID            string  `json:"-"`
	PaymentClientReferenceID      string  `json:"-"`
	PaymentChargeID               string  `json:"-"`
	PaymentChargeAmount           float64 `json:"-"`
	PaymentFeeID                  string  `json:"-"`
	RefundOfPaymentFeeAmount      float64 `json:"-"`
	PaymentChargeSettlementStatus string  `json:"-"`

	// Refund object
	*Refund

	// LedgerRefundObject
	RefundLedgerID          string  `json:"-"`
	RefundLedgerReasonType  *string `json:"-"`
	RefundLedgerReasonDesc  *string `json:"-"`
	RefundLedgerReferenceID string  `json:"-"`
}

type FilterRefundRequest struct {
	MerchantID        string     `json:"-"`
	UUID              string     `json:"uuid"`
	PaymentSessionID  string     `json:"paymentSessionId"`
	ChargeID          string     `json:"chargeId"`
	ClientReferenceID string     `json:"clientReferenceId"`
	Status            string     `json:"status"`
	StartCreatedAt    *time.Time `json:"startCreatedAt"`
	EndCreatedAt      *time.Time `json:"endCreatedAt"`

	Sort    string `json:"sort"`
	SortBy  string `json:"sortBy"`
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
}

type GetExistingRefundListRequest struct {
	PaymentID string
	Status    string
}

type ListByPaymentIDRequest struct{ Status string }
