package flipProcessorModel

import (
	"github.com/paper-indonesia/pdk/go/snap"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	flipProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/flipProcessor/common"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/accountInquiry"
)

type AccountInquiryRequest struct {
	AccountNo     string `json:"account_number"`
	BankCode      string `json:"bank_code"`
	AccountHolder string `json:"account_holder"`
	Status        string `json:"status"`
	InquiryKey    string `json:"inquiry_key"`
}

func (a *AccountInquiryRequest) ToAccountInquiryResponse() *routingProcessorModel.InquiryAccountResponseData {
	snapStatus := snap.SNAP_INPROGRESS
	switch a.Status {
	case constant.FlipAccountInquiryStatusSuccess:
		snapStatus = snap.SNAP_SUCCESS
	case constant.FlipAccountInquiryStatusBlacklisted,
		constant.FlipAccountInquiryStatusInvalid,
		constant.FlipAccountInquiryStatusSuspectedAccount:
		snapStatus = snap.SNAP_INVALID_ACCOUNT
	}

	responseCode, responseMessage := snap.GenerateResponseCode(snapStatus, snap.SNAP_SERVICE_ACCOUNT_INQUIRY_EXTERNAL)
	return &routingProcessorModel.InquiryAccountResponseData{
		ResponseCode:           responseCode,
		ResponseMessage:        responseMessage,
		PartnerReferenceNo:     a.InquiryKey,
		BeneficiaryAccountName: a.AccountHolder,
		BeneficiaryAccountNo:   a.AccountNo,
		BeneficiaryBankCode:    flipProcessorModel.GetBankCode(a.BankCode),
		ProcessorReference:     constant.FlipPGProcessor,
		Status:                 a.Status,
	}
}
