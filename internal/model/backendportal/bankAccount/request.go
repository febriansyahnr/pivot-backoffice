package bankAccount

type BankAccountRequest struct {
	UUID                   string `json:"uuid" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	MerchantID             string `json:"merchantId" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" example:"8000800808" validate:"required,numeric"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" example:"Yories Yolanda" validate:"required"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode" example:"008" validate:"required,numeric"`
	BeneficiaryBankName    string `json:"beneficiaryBankName" example:"Bank 008" validate:"required"`
}

type CreateBankAccountRequest struct {
	UUID                   string `json:"uuid" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	MerchantID             string `json:"merchantId" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" example:"8000800808" validate:"required,numeric"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" example:"Yories Yolanda" validate:"required"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode" example:"008" validate:"required,numeric"`
	BeneficiaryBankName    string `json:"beneficiaryBankName" example:"Bank 008" validate:"required"`
	CreatedBy              string `json:"createdBy" validate:"required"`
}

type UpdateBankAccountRequest struct {
	UUID                   string `json:"uuid" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	MerchantID             string `json:"merchantId" example:"a1a1a1a1-1a1a-1a1a-1a1a-1a1a1a1a1a1a"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" example:"8000800808" validate:"required,numeric"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" example:"Yories Yolanda" validate:"required"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode" example:"008" validate:"required,numeric"`
	BeneficiaryBankName    string `json:"beneficiaryBankName" example:"Bank 008" validate:"required"`
	UpdatedBy              string `json:"updatedBy" validate:"required"`
}

type GetMerchantBankAccountRequest struct {
	MerchantID          string
	RequesterMerchantID string
}
