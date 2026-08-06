package snapCoreModel

type InquiryAccountRequest struct {
	BeneficiaryBankCode    string         `json:"beneficiaryBankCode" validate:"required" example:"013"`         //required. contains bank code of beneficiary
	BeneficiaryAccountNo   string         `json:"beneficiaryAccountNo" validate:"required" example:"1234567890"` //required. contains account number of beneficiary
	BeneficiaryAccountName string         `json:"beneficiaryAccountName" example:"Sdr. Asep"`
	PartnerReferenceNo     string         `json:"partnerReferenceNo" example:"BT-123"` //optional. When nil, will be generated
	MerchantID             string         `json:"-"`
	AdditionalInfo         map[string]any `json:"additionalInfo"`
}
