package cardFundedPayoutModel

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vendor"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type CreateSavedCardRequest struct {
	ReferenceID string `json:"referenceId" validate:"required"`

	MerchantID string `json:"-"`
	CreatedBy  string `json:"-"`
}

type FilterGetSavedCardList struct {
	MerchantID string `json:"-"`

	StartCreatedAt *time.Time `json:"-"`
	EndCreatedAt   *time.Time `json:"-"`
	Sort           string     `json:"-"`
	SortBy         string     `json:"-"`
	Page           int        `json:"-"`
	PerPage        int        `json:"-"`
}

type PayoutActionRequest struct {
	MerchantID string `json:"-" validate:"required,uuid"`
	UserID     string `json:"-" validate:"required,uuid"`
	UserName   string `json:"-" validate:"required"`
}

func (p *PayoutActionRequest) SetMerchantID(id string) { p.MerchantID = id }
func (p *PayoutActionRequest) SetUserID(id string)     { p.UserID = id }
func (p *PayoutActionRequest) SetUserName(name string) { p.UserName = name }

type CreatePayoutRequest struct {
	VendorID         string               `json:"vendorId" validate:"required,uuid"`
	ReferenceID      string               `json:"referenceId" validate:"required,min=1,max=100"`
	Amount           common.AmountRequest `json:"amount" validate:"required"`
	Remarks          string               `json:"remarks" validate:"max=255"`
	SettlementMethod string               `json:"settlementMethod" validate:"required,oneof=STANDARD"`
	CardID           string               `json:"cardId" validate:"required,uuid"`
	PayoutActionRequest
}

func (c *CreatePayoutRequest) ToCreateDisbursement(vendor *vendor.Vendor, card *GetSavedCardResponse, fee *feeModel.FeeMetadataObject) disbursementModel.Disbursement {
	metadata := disbursementModel.Metadata{
		CardFundedDetail: &disbursementModel.CardFundedDetailMetadata{
			VendorID:   vendor.UUID,
			VendorName: vendor.Name,
			Card: &disbursementModel.CardFundedDetailMetadataCard{
				ID:             card.ID,
				CardName:       card.CardName,
				PaymentChannel: card.PaymentChannel,
				IssuingBank:    card.IssuingBank,
				Last4Digits:    card.Last4,
				ExpiryMonth:    card.ExpiryMonth,
				ExpiryYear:     card.ExpiryYear,
			},
			SettlementMethod: c.SettlementMethod,
		},
		FeeDetail: *fee,
	}
	rawMetadata, _ := json.Marshal(metadata)

	now := time.Now().UTC()
	return disbursementModel.Disbursement{
		UUID:                   util.GenerateUUID().String(),
		ReferenceID:            c.ReferenceID,
		MerchantID:             c.MerchantID,
		SenderName:             card.MerchantName,
		BeneficiaryBankCode:    vendor.BankCode,
		BeneficiaryBankName:    util.ValueToPtr(vendor.BankName),
		BeneficiaryAccountNo:   vendor.AccountNumber,
		BeneficiaryAccountName: vendor.AccountName,
		Currency:               c.Amount.Currency,
		Amount:                 decimal.NewFromFloat(c.Amount.Value),
		Fee:                    util.ValueToPtr(decimal.NewFromFloat(fee.FinalAmount)),
		TotalAmount:            decimal.NewFromFloat(c.Amount.Value + fee.FinalAmount),
		Status:                 constant.DisbursementStatusWaiting,
		Remark:                 util.ValueToPtr(c.Remarks),
		Metadata:               types.NullJSONText{Valid: true, JSONText: rawMetadata},
		CreatedFrom:            util.ValueToPtr(constant.SourceDashboard),
		CreatedBy:              util.ValueToPtr(c.UserID),
		CreatedAt:              now,
		UpdatedAt:              now,
		MetadataObj:            metadata,
	}
}

type ApprovePayoutRequest struct {
	ID  string `json:"-" validate:"required,uuid"`
	CVC string `json:"cvc" validate:"required,numeric,min=3,max=4"`
	PayoutActionRequest
	// For internal use
	CardID                string  `json:"-" validate:"-"`
	CardName              string  `json:"-" validate:"-"`
	CardToken             string  `json:"-" validate:"-"`
	VendorID              string  `json:"-" validate:"-"`
	VendorName            string  `json:"-" validate:"-"`
	SettlementMethod      string  `json:"-" validate:"-"`
	ActiveProcessor       string  `json:"-" validate:"-"` // MPGS / CYBS
	ProcessorLimit        float64 `json:"-" validate:"-"`
	CitAcquirerMerchantID string  `json:"-" validate:"-"`
	MitAcquirerMerchantID string  `json:"-" validate:"-"`
	// Payout details
	Payout *disbursementModel.Disbursement `json:"-" validate:"-"`
}

type RejectPayoutRequest struct {
	ID     string `json:"-" validate:"required,uuid"`
	Reason string `json:"reason" validate:"required,max=255"`
	PayoutActionRequest
}

type GetSavedCardDetailRequest struct {
	MerchantID string `json:"-"`
	CardID     string `json:"-"`
}

type FilterGetPayoutList struct {
	MerchantID        string     `json:"-"`
	StartCreatedAt    *time.Time `json:"startDate"`
	EndCreatedAt      *time.Time `json:"endDate"`
	TransactionStatus string     `json:"transactionStatus"` // PROCESSING/SUCCESS/FAILED
	ApprovalStatus    string     `json:"approval"`          // WAITING/APPROVED/REJECTED
	SearchID          string     `json:"-"`                 // search by payout ID or reference ID
	Sort              string     `json:"-"`
	SortBy            string     `json:"-"`
	Page              int64      `json:"-"`
	PerPage           int64      `json:"-"`
}

// GetPayoutDetailRequest is the filter for getting payout detail
type GetPayoutDetailRequest struct {
	PayoutID   string `json:"-"`
	MerchantID string `json:"-"`
}

// FilterGetPayoutInsights is the filter for getting payout insights.
type FilterGetPayoutInsights struct {
	MerchantID     string     `json:"-"`
	StartCreatedAt *time.Time `json:"-"`
	EndCreatedAt   *time.Time `json:"-"`
}

type ProcessFinishCardFundedPayoutSettlementRequest struct {
	MerchantID  string
	ReferenceID string
}

type ExecuteSubsequentPaymentRequest struct {
	MerchantID  string
	ReferenceID string
	VendorID    string
	VendorName  string
}

type GetPayoutTransactionListRequest struct {
	MerchantID    string    `json:"merchantId" validate:"required,uuid"`
	StrStartDate  string    `json:"-" validate:"required,datetime=2006-01-02T15:04:05Z"`
	StrEndDate    string    `json:"-" validate:"required,datetime=2006-01-02T15:04:05Z"`
	StartDate     time.Time `json:"startDate" validate:"-"`
	EndDate       time.Time `json:"endDate" validate:"-"`
	TrxStatus     string    `json:"trxStatus" validate:"omitempty,oneof=INITIATE SCHEDULED PENDING SUCCESS FAILED"`
	TrxReasonType string    `json:"trxReasonType" validate:"omitempty,uppercase"`
}

type PatchPayoutTransactionStatusRequest struct {
	PayoutID          string  `json:"-" validate:"required,uuid"`
	Status            string  `json:"status" validate:"required,oneof=SUCCESS FAILED"`
	BankReferenceNo   string  `json:"bankReferenceNo" validate:"required_if=Status SUCCESS"`
	ReconReferenceNo  string  `json:"reconReferenceNo" validate:"required_if=Status SUCCESS"`
	ReasonType        *string `json:"reasonType" validate:"required_if=Status FAILED,omitempty,uppercase"`
	ReasonDescription *string `json:"reasonDescription" validate:"required_if=Status FAILED"`
}

func (r *ApprovePayoutRequest) PrepareCfpProcessor(cfpConfig *paymentMethodModel.CardFundedPayoutConfig, partnerConfig *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest, processorLimitDefault float64) error {
	processorLimit := processorLimitDefault
	if processor, ok := cfpConfig.Processors[cfpConfig.ActiveProcessor]; ok {
		processorLimit = processor.Limit
	}
	r.ProcessorLimit = processorLimit

	for _, cnf := range partnerConfig.Card.Items {
		if cnf.PartnerProcessor != cfpConfig.ActiveProcessor {
			continue
		}
		switch cnf.CardFundedPayoutType {
		case constant.CardTransactionTypeCIT:
			r.CitAcquirerMerchantID = cnf.AcquirerMerchantID

		case constant.CardTransactionTypeMIT:
			r.MitAcquirerMerchantID = cnf.AcquirerMerchantID
		}
	}

	invalidCfpConfig := r.CitAcquirerMerchantID == "" ||
		r.MitAcquirerMerchantID == "" ||
		r.ProcessorLimit == 0
	if invalidCfpConfig {
		return pkgErr.New(response.HttpErrRequest, errors.New("invalid configuration for card-funded payout"))
	}
	return nil
}
