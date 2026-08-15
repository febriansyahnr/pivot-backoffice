package fdscommon

import (
	"time"

	common "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"

	"github.com/shopspring/decimal"
)

// FdsRiskAssessment represents the risk assessment data to be stored in ledger additional_info
type FdsRiskAssessment struct {
	Score            decimal.Decimal `json:"score"`
	Level            string          `json:"level"`          // from ruleEvaluation.result
	Recommendation   string          `json:"recommendation"` // Approve/Reject based on status
	Status           string          `json:"status"`         // PASSED/REJECTED/NOT_EVALUATED
	EvaluatedAt      time.Time       `json:"evaluatedAt"`
	IsFraud          *bool           `json:"isFraud,omitempty"`
	ChargebackStatus string          `json:"chargebackStatus,omitempty"`
	ChargebackNotes  string          `json:"chargebackNotes,omitempty"`
}

func (f *FdsRiskAssessment) Update(updateFds *FdsRiskAssessment) {
	if !updateFds.Score.IsZero() {
		f.Score = updateFds.Score
	}

	if updateFds.Level != "" {
		f.Level = updateFds.Level
	}

	if updateFds.Recommendation != "" {
		f.Recommendation = updateFds.Recommendation
	}

	if updateFds.Status != "" {
		f.Status = updateFds.Status
	}

	if !updateFds.EvaluatedAt.IsZero() {
		f.EvaluatedAt = updateFds.EvaluatedAt
	}

	if updateFds.ChargebackStatus != "" {
		f.ChargebackStatus = updateFds.ChargebackStatus
	}

	if updateFds.IsFraud != nil {
		f.IsFraud = updateFds.IsFraud
	}
}

type Merchant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	RiskLevel string `json:"riskLevel"`
}

type Transaction struct {
	ID                string         `json:"id"`
	ClientReferenceID string         `json:"clientReferenceId"`
	Type              string         `json:"type"`
	Amount            common.Amount2 `json:"amount"`
	CreatedAt         time.Time      `json:"createdAt"`
	UpdatedAt         time.Time      `json:"updatedAt"`
	CreatedFrom       string         `json:"createdFrom,omitempty"`
	Status            string         `json:"status,omitempty"`
	PaymentStatus     string         `json:"paymentStatus,omitempty"`

	// Additional Information for Chargeback
	ChargebackInfo
}

type ChargebackInfo struct {
	IsFraud          *bool  `json:"isFraud,omitempty"`
	ChargebackStatus string `json:"chargebackStatus,omitempty"`
	ChargebackNotes  string `json:"chargebackNotes,omitempty"`
}

type PayoutDestination struct {
	BankCode      string `json:"bankCode"`
	AccountNumber string `json:"accountNumber"`
	AccountName   string `json:"accountName"`
}

type RiskAssessmentResult interface {
	RawResult() []byte
}

type PaymentMethod struct {
	Type string                 `json:"type"`
	Card *PaymentMethodTypeCard `json:"card,omitempty"`
}

type PaymentMethodTypeCard struct {
	BankMerchantID  string `json:"bankMerchantId"`
	AcquirerName    string `json:"acquirerName"`
	ThreeDsMethod   string `json:"threeDsMethod"`
	CardFingerprint string `json:"cardFingerprint"`
	CardNumber      string `json:"cardNumber"`
	Last4           int    `json:"last4"`
	First6          int    `json:"first6"`
	CardBrand       string `json:"cardBrand"`
	CardCountryCode string `json:"cardCountryCode"`
	CardType        string `json:"cardType"`
	IssuerName      string `json:"issuerName"`
	ECICode         string `json:"eciCode"`
	ApprovalCode    string `json:"approvalCode,omitempty"`
	CvvCode         string `json:"cvvCode,omitempty"`
}
