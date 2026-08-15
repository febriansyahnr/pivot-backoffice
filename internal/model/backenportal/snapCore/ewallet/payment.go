package ewallet

import (
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
)

type EwalletPaymentRequest struct {
	OriginalReferenceId string                       `json:"originalReferenceId"`
	Acquirer            string                       `json:"acquirer"`
	MerchantId          string                       `json:"merchantId,omitempty"`
	ExternalStoreId     string                       `json:"externalStoreId,omitempty"`
	SubMerchantId       string                       `json:"subMerchantId,omitempty"`
	Amount              commonModel.Amount           `json:"amount"`
	ValidUpTo           string                       `json:"validUpTo"`
	UrlParams           []snapCoreModel.SnapURLParam `json:"urlParams"`
	AdditionalInfo      map[string]any               `json:"additionalInfo"`
}

type SnapCoreEwalletPaymentLinkResponse struct {
	*snapCoreModel.StandardResponse
	Data *EwalletPaymentLinkResponse `json:"data"`
}

type EwalletPaymentLinkResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	UUID               string `json:"uuid"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	WebRedirectionURL  string `json:"webRedirectUrl"`
	AppRedirectionURL  string `json:"appRedirectUrl"`
}
