package routingProcessorModel

type InquiryAccountRequest struct {
	MerchantID             string         `json:"merchantId" validate:"required,uuid"`
	BeneficiaryBankCode    string         `json:"beneficiaryBankCode" validate:"required" example:"013"`         //required. contains bank code of beneficiary
	BeneficiaryAccountNo   string         `json:"beneficiaryAccountNo" validate:"required" example:"1234567890"` //required. contains account number of beneficiary
	BeneficiaryAccountName string         `json:"beneficiaryAccountName" example:"Bank 013"`
	PartnerReferenceNo     string         `json:"partnerReferenceNo" example:"BT-123"` //optional. When nil, will be generated
	AdditionalInfo         map[string]any `json:"additionalInfo"`
}
