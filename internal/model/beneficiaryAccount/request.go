package beneficiaryAccountModel

import "time"

type CheckAccountRequest struct {
	BeneficiaryAccountNo string         `json:"beneficiaryAccountNo" validate:"required,numeric" example:"8000800808"`
	BeneficiaryBankCode  string         `json:"beneficiaryBankCode" validate:"required" example:"008"`
	MerchantID           string         `json:"-" example:"MID-123"`
	AdditionalInfo       map[string]any `json:"additionalInfo,omitempty"`
}

type GetBeneficiaryAccountFilterRequest struct {
	MerchantID             string     `json:"merchantId"`
	BeneficiaryAccountNo   string     `json:"beneficiaryAccountNo"`
	BeneficiaryAccountName string     `json:"beneficiaryAccountName"`
	StartCreatedAt         *time.Time `json:"startCreatedAt"`
	EndCreatedAt           *time.Time `json:"endCreatedAt"`
	IsXb                   bool       `json:"isXb"`
}
