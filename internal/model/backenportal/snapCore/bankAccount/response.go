package snapCoreModel

type InquiryAccountResponse struct {
	Data    InquiryAccountResponseData `json:"data"`
	Code    string                     `json:"code"`
	Message string                     `json:"message"`
	Error   interface{}                `json:"error,omitempty"`
}

type InquiryAccountResponseData struct {
	ResponseCode           string `json:"responseCode" example:"200xx200"`
	ResponseMessage        string `json:"responseMessage" example:"Success"`
	PartnerReferenceNo     string `json:"partnerReferenceNo" example:"BT-120"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" example:"John Doe"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" example:"1234567890"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode,omitempty" example:"013"`
	BeneficiaryBankName    string `json:"beneficiaryBankName,omitempty" example:"PERMATA"`
	IsVirtualAccount       bool   `json:"isVirtualAccount" example:"false"`
}
