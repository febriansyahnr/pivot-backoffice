package paymentModel

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
)

type GetInvestigatedPaymentsFilterRequest struct {
	InvestigationStatus string     `json:"investigationStatus"`
	FromDate            *time.Time `json:"fromDate"`
	ToDate              *time.Time `json:"toDate"`
	PaymentReferenceID  string     `json:"paymentReferenceId"`
	MerchantID          string     `json:"merchantId"`
	PaymentMethod       string     `json:"paymentMethod"` // Payment method type (VIRTUAL_ACCOUNT, QRIS, etc)
	Channel             string     `json:"channel"`       // Payment channel/bank name
	Page                int        `json:"page"`
	Limit               int        `json:"limit"`
	SortBy              string     `json:"sortBy"`
	Sort                string     `json:"sort"`
}

type InvestigatedPaymentResponse struct {
	UUID                string          `json:"uuid"`
	PaymentReferenceID  string          `json:"paymentReferenceId"`
	Amount              decimal.Decimal `json:"amount"`
	Currency            string          `json:"currency"`
	MerchantID          string          `json:"merchantId"`
	MerchantName        string          `json:"merchantName"`
	PaymentMethod       string          `json:"paymentMethod"`
	PaymentChannel      string          `json:"paymentChannel"`
	PaymentStatus       string          `json:"paymentStatus"`
	InvestigationStatus string          `json:"investigationStatus"`
	StartedAt           *time.Time      `json:"startedAt"`
	CompletedAt         *time.Time      `json:"completedAt"`
	LastUpdatedAt       time.Time       `json:"lastUpdatedAt"`
	Notes               *string         `json:"notes"`
}

type UpdateInvestigationRequest struct {
	InvestigationStatus string  `json:"investigationStatus" validate:"required,oneof=INVESTIGATION_SUCCESS INVESTIGATION_FAILED"`
	Notes               *string `json:"notes" validate:"omitempty,max=200"`
}

type UpdateInvestigationStatusRequest struct {
	PaymentID   string
	Status      string
	Notes       *string
	CompletedAt time.Time
}

type UpdateInvestigationResponse struct {
	PaymentReferenceID  string     `json:"paymentReferenceId"`
	InvestigationStatus string     `json:"investigationStatus"`
	CompletedAt         *time.Time `json:"completedAt"`
	LastUpdatedAt       time.Time  `json:"lastUpdatedAt"`
	Notes               *string    `json:"notes"`
}

type InvestigatedPaymentDTO struct {
	UUID              string          `db:"uuid"`
	ReferenceId       string          `db:"reference_id"`
	Amount            decimal.Decimal `db:"amount"`
	Currency          string          `db:"currency"`
	MerchantID        string          `db:"merchant_id"`
	MerchantName      string          `db:"merchant_name"`
	PaymentMethodType string          `db:"payment_method_type"`
	PaymentChannel    string          `db:"payment_channel"`
	Status            string          `db:"status"`
	ReasonType        string          `db:"reason_type"`
	ReasonDescription *string         `db:"reason_description"`
	UpdatedAt         time.Time       `db:"updated_at"`
	StartedAt         *time.Time      `db:"started_at"`
	CompletedAt       *time.Time      `db:"completed_at"`
}

func (dto *InvestigatedPaymentDTO) ToResponse() *InvestigatedPaymentResponse {
	return &InvestigatedPaymentResponse{
		UUID:                dto.UUID,
		PaymentReferenceID:  dto.ReferenceId,
		Amount:              dto.Amount,
		Currency:            dto.Currency,
		MerchantID:          dto.MerchantID,
		MerchantName:        dto.MerchantName,
		PaymentMethod:       dto.PaymentMethodType,
		PaymentChannel:      dto.PaymentChannel,
		PaymentStatus:       dto.Status,
		InvestigationStatus: dto.ReasonType,
		StartedAt:           dto.StartedAt,
		LastUpdatedAt:       dto.UpdatedAt,
		CompletedAt:         dto.CompletedAt,
		Notes:               dto.ReasonDescription,
	}
}

type MonthlyReconciliationRequest struct {
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type PaymentInvestigationMonthlyReconciliation struct {
	UUID                   string         `db:"uuid"`
	Date                   time.Time      `db:"date"`
	MerchantID             string         `db:"merchant_id"`
	RawPaymentIDs          types.JSONText `db:"payment_ids"`
	PaymentIDs             []string       `db:"-"`
	PaymentCount           int            `db:"payment_count"`
	GrossAmount            float64        `db:"gross_amount"`
	FeeAmount              float64        `db:"fee_amount"`
	NetAmount              float64        `db:"net_amount"`
	PlatformLossPercentage float64        `db:"platform_loss_percentage"`
	PlatformMaxLoss        float64        `db:"platform_max_loss"`
	PlatformLossAmount     float64        `db:"platform_loss_amount"`
	MerchantLossAmount     float64        `db:"merchant_loss_amount"`
	CreatedAt              time.Time      `db:"created_at"`
}

type CalculateInvestigationMonthlyReconciliation struct {
	MerchantID             string         `db:"merchant_id" json:"merchantId"`
	RawPaymentIDs          types.JSONText `db:"payment_ids" json:"-"`
	PaymentIDs             []string       `db:"-" json:"paymentIds"`
	PaymentCount           int            `db:"payment_count" json:"paymentCount"`
	GrossAmount            float64        `db:"gross_amount" json:"grossAmount"`
	FeeAmount              float64        `db:"fee_amount" json:"feeAmount"`
	NetAmount              float64        `db:"net_amount" json:"netAmount"`
	PlatformLossPercentage float64        `db:"platform_loss_percentage" json:"platformLossPercentage"`
	PlatformMaxLoss        float64        `db:"platform_max_loss" json:"platformMaxLoss"`
	PlatformLossAmount     float64        `db:"platform_loss_amount" json:"platformLossAmount"`
	MerchantLossAmount     float64        `db:"merchant_loss_amount" json:"merchantLossAmount"`
}

func (c *CalculateInvestigationMonthlyReconciliation) ToCreateAccountTransactionRequest(referenceID string, date time.Time) *orchestrator_model.CreateAccountTransactionRequest {
	return &orchestrator_model.CreateAccountTransactionRequest{
		UUID:                 util.GenerateUUID(),
		ReferenceID:          referenceID,
		MerchantID:           util.ParseUUID(c.MerchantID),
		Currency:             constant.CurrencyIDR,
		Debit:                c.MerchantLossAmount,
		Type:                 constant.TypePayment,
		Channel:              constant.ChannelInvestigation,
		Status:               constant.StatusSuccess,
		ReasonType:           util.ValueToPtr(constant.TypeFinalFailedDeduction),
		TransactionTimestamp: date,
		Usecase:              constant.TypePayment,
		SettlementModel:      util.ValueToPtr(constant.SettlementModelAggregator),
	}
}

func (c *CalculateInvestigationMonthlyReconciliation) ToPaymentInvestigationMonthlyReconciliation(id string, date time.Time) PaymentInvestigationMonthlyReconciliation {
	return PaymentInvestigationMonthlyReconciliation{
		UUID:                   id,
		Date:                   date,
		MerchantID:             c.MerchantID,
		RawPaymentIDs:          c.RawPaymentIDs,
		PaymentIDs:             c.PaymentIDs,
		PaymentCount:           c.PaymentCount,
		GrossAmount:            c.GrossAmount,
		FeeAmount:              c.FeeAmount,
		NetAmount:              c.NetAmount,
		PlatformLossPercentage: c.PlatformLossPercentage,
		PlatformMaxLoss:        c.PlatformMaxLoss,
		PlatformLossAmount:     c.PlatformLossAmount,
		MerchantLossAmount:     c.MerchantLossAmount,
		CreatedAt:              time.Now().UTC(),
	}
}

type InvestigationDownloadHistoryRequest struct {
	MerchantId          string     `json:"-" validate:"required,uuid"`
	InvestigationStatus string     `json:"investigationStatus" validate:"omitempty,oneof=INVESTIGATION_IN_PROCESS INVESTIGATION_SUCCESS INVESTIGATION_FAILED"`
	PaymentReferenceID  string     `json:"paymentReferenceId" validate:"-"`
	PaymentMethod       string     `json:"paymentMethod" validate:"omitempty,oneof=QRIS VIRTUAL_ACCOUNT CREDIT_CARD EWALLET"`
	Channel             string     `json:"channel" validate:"-"`
	FromDate            *time.Time `json:"fromDate" validate:"omitempty"`
	ToDate              *time.Time `json:"toDate" validate:"omitempty"`
}

type InvestigationDownloadHistoryResponse struct {
	URL string `json:"url"`
}

type SummaryRow struct {
	TotalInProgress float64 `db:"total_in_progress"`
	TotalSuccess    float64 `db:"total_success"`
	TotalFailed     float64 `db:"total_failed"`
	Currency        string  `db:"currency"`
}

type GetInvestigationProofOfPaymentRequest struct {
	PaymentID string `json:"paymentId"`
}

type GetInvestigationProofOfPaymentResponse struct {
	SignedURL     string    `json:"signedUrl"`
	ExpiresAt     time.Time `json:"expiresAt"`
	MerchantNotes string    `json:"merchantNotes"`
}
