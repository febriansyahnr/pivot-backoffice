package snapCoreModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	common "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
)

type RegUploadResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Id      string `json:"uuid"`
		MediaId string `json:"mediaId"`
	} `json:"data,omitempty"`
	Error interface{} `json:"error,omitempty"`
}

type UploadDocumentResp struct {
	Id      string
	MediaId string
}

type CancelQrMpmResponse struct {
	*common.StandardResponse
	Data *CancelQrMpmResponseData `json:"data"`
}

type CancelQrMpmResponseData struct {
	UUID               string `json:"uuid"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	Status             string `json:"status"`
	QrType             string `json:"qrType"`
}

type RefundQRMPMResponse struct {
	*common.StandardResponse
	Data *RefundResponseData `json:"data"`
}

type RefundResponseData struct {
	OriginalReferenceNo        string                 `json:"originalReferenceNo,omitempty"`
	RefundNo                   string                 `json:"refundNo,omitempty"`
	OriginalExternalID         string                 `json:"originalExternalId,omitempty"`
	RefundTime                 time.Time              `json:"refundTime,omitempty"`
	AdditionalInfo             map[string]interface{} `json:"additionalInfo,omitempty"`
	OriginalPartnerReferenceNo string                 `json:"originalPartnerReferenceNo,omitempty"`
	ResponseMessage            string                 `json:"responseMessage"`
	PartnerRefundNo            string                 `json:"partnerRefundNo,omitempty"`
	ResponseCode               string                 `json:"responseCode"`
	RefundAmount               commonModel.Amount     `json:"refundAmount,omitempty"`
}
