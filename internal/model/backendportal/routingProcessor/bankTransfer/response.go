package routingProcessorModel

import (
	"encoding/json"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankTransfer"
)

type BankTransferResponseData struct {
	ResponseCode            string             `json:"responseCode"`
	ResponseMessage         string             `json:"responseMessage"`
	UUID                    string             `json:"uuid"`
	PartnerReferenceNo      string             `json:"partnerReferenceNo,omitempty"`
	BankReferenceNo         string             `json:"bankReferenceNo,omitempty"`
	BankProcessor           string             `json:"bankProcessor,omitempty"`
	Amount                  commonModel.Amount `json:"amount,omitempty"`
	BeneficiaryAccountNo    string             `json:"beneficiaryAccountNo,omitempty"`
	BeneficiaryBankCode     string             `json:"beneficiaryBankCode,omitempty"`
	SourceAccountNo         string             `json:"sourceAccountNo,omitempty"`
	Status                  string             `json:"status"`
	TransferType            string             `json:"transferType"`
	ExternalID              string             `json:"externalId,omitempty"`
	Reason                  string             `json:"reason,omitempty"`
	ProcessorReference      string             `json:"processorReference,omitempty"`
	Metadata                map[string]any     `json:"metadata,omitempty"`
	TransactionDate         time.Time          `json:"transactionDate,omitempty"`
	LatestTransactionStatus string             `json:"latestTransactionStatus,omitempty"`
	AdditionalInfo          map[string]any     `json:"additionalInfo,omitempty"`
	// Internal Used
	Transaction *orchestrator_model.AccountTransactionWithUseCase `json:"-"`
}

func (d *BankTransferResponseData) ToSnapBankTransferResponseData() snapCoreModel.BankTransferResponseData {
	return snapCoreModel.BankTransferResponseData{
		ResponseCode:         d.ResponseCode,
		ResponseMessage:      d.ResponseMessage,
		UUID:                 d.UUID,
		PartnerReferenceNo:   d.PartnerReferenceNo,
		BankReferenceNo:      d.BankReferenceNo,
		BankProcessor:        d.BankProcessor,
		Amount:               d.Amount,
		BeneficiaryAccountNo: d.BeneficiaryAccountNo,
		BeneficiaryBankCode:  d.BeneficiaryBankCode,
		SourceAccountNo:      d.SourceAccountNo,
		Status:               d.Status,
		TransferType:         d.TransferType,
		ExternalID:           d.ExternalID,
		TransactionDate:      d.TransactionDate,
	}
}

type InquiryStatusResponse struct {
	ResponseCode               string             `json:"responseCode"`
	ResponseMessage            string             `json:"responseMessage"`
	OriginalReferenceNo        string             `json:"originalReferenceNo"`
	OriginalPartnerReferenceNo string             `json:"originalPartnerReferenceNo"`
	OriginalExternalID         string             `json:"originalExternalId"`
	ServiceCode                string             `json:"serviceCode"`
	TransactionDate            string             `json:"transactionDate,omitempty"`
	Amount                     commonModel.Amount `json:"amount,omitempty"`
	BeneficiaryAccountNo       string             `json:"beneficiaryAccountNo"`
	BeneficiaryBankCode        string             `json:"beneficiaryBankCode"`
	Currency                   string             `json:"currency"`
	PreviousResponseCode       string             `json:"previousResponseCode"`
	ReferenceNumber            string             `json:"referenceNumber"`
	SourceAccountNo            string             `json:"sourceAccountNo"`
	TransactionID              string             `json:"transactionId"`
	LatestTransactionStatus    string             `json:"latestTransactionStatus"`
	TransactionStatusDesc      string             `json:"transactionStatusDesc"`
	AdditionalInfo             map[string]any     `json:"additionalInfo"`
}

func (r *BankTransferResponseData) GetReconReferenceNo() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		return ""
	}

	var snapCoreResp snapCoreModel.BankTransferResponseData
	if err = json.Unmarshal(bytes, &snapCoreResp); err != nil {
		return ""
	}

	return snapCoreResp.GetReconReferenceNo()
}
