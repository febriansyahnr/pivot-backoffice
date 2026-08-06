package beneficiaryAccountModel

import "time"

type CheckAccountResponse struct {
	UUID                   string                 `json:"uuid" example:"b3b3b3b3-3b3b-3b3b-3b3b-3b3b3b3b3b3b"`
	BeneficiaryAccountNo   string                 `json:"beneficiaryAccountNo" example:"8000800808"`
	BeneficiaryAccountName string                 `json:"beneficiaryAccountName" example:"Yories Yolanda"`
	BeneficiaryBankCode    string                 `json:"beneficiaryBankCode" example:"008"`
	BeneficiaryBankName    string                 `json:"beneficiaryBankName" example:"Bank 008"`
	CreatedAt              time.Time              `json:"createdAt" example:"2021-01-01T00:00:00Z"`
	UpdatedAt              time.Time              `json:"updatedAt" example:"2021-01-01T00:00:00Z"`
	AdditionalInfo         *AccountAdditionalInfo `json:"additionalInfo,omitempty"`
}
