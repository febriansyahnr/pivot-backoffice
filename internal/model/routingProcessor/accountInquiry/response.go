package routingProcessorModel

import snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"

type InquiryAccountResponseData struct {
	ResponseCode           string `json:"responseCode" example:"200xx200"`
	ResponseMessage        string `json:"responseMessage" example:"Success"`
	PartnerReferenceNo     string `json:"partnerReferenceNo" example:"BT-120"`
	BeneficiaryAccountName string `json:"beneficiaryAccountName" example:"John Doe"`
	BeneficiaryAccountNo   string `json:"beneficiaryAccountNo" example:"1234567890"`
	BeneficiaryBankCode    string `json:"beneficiaryBankCode,omitempty" example:"013"`
	BeneficiaryBankName    string `json:"beneficiaryBankName,omitempty" example:"PERMATA"`
	ProcessorReference     string `json:"-"`
	Status                 string `json:"status,omitempty"`
	IsVirtualAccount       bool   `json:"isVirtualAccount,omitempty"`
}

func (inq *InquiryAccountResponseData) ToSnapCoreResponseData() *snapCoreModel.InquiryAccountResponseData {
	return &snapCoreModel.InquiryAccountResponseData{
		ResponseCode:           inq.ResponseCode,
		ResponseMessage:        inq.ResponseMessage,
		PartnerReferenceNo:     inq.PartnerReferenceNo,
		BeneficiaryAccountName: inq.BeneficiaryAccountName,
		BeneficiaryAccountNo:   inq.BeneficiaryAccountNo,
		BeneficiaryBankCode:    inq.BeneficiaryBankCode,
		BeneficiaryBankName:    inq.BeneficiaryBankName,
		IsVirtualAccount:       inq.IsVirtualAccount,
	}
}
