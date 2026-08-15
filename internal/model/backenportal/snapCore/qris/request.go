package snapCoreModel

import commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"

type UploadDocumentReq struct {
	Acquirer       string `form:"acquirer"`
	RegistrationId string `form:"registrationId"`
	DocumentType   string `form:"documentType"`
	DocumentNumber string `form:"number"`
	ObjectName     string `form:"-"`
	RawFile        []byte `form:"-"`
}

type QrMpmPaymentSimulationRequest struct {
	PartnerReferenceNo string             `json:"partnerReferenceNo"`
	Status             string             `json:"status"`
	Amount             commonModel.Amount `json:"amount"`
}

type QRMPMRefundRequest struct {
	QRID           string                 `json:"paymentId"`
	Reason         string                 `json:"refundReason"`
	AdditionalInfo map[string]interface{} `json:"additionalInfo"`
	Amount         commonModel.Amount     `json:"refundAmount"`
}
