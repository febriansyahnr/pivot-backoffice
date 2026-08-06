package snapCoreModel

import (
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
)

const DefaultPurchaseOfTransaction = "99"

type BankTransferRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	BTBeneficiaryRequest
	Amount               commonModel.Amount       `json:"amount"`
	Currency             string                   `json:"currency,omitempty"`
	Remark               string                   `json:"remark,omitempty"`
	PurposeOfTransaction string                   `json:"purposeOfTransaction,omitempty"`
	SourceAccountNo      string                   `json:"sourceAccountNo"`
	SourceAccountName    string                   `json:"sourceAccountName"`
	TransactionDate      time.Time                `json:"transactionDate"`
	AdditionalInfo       map[string]any           `json:"additionalInfo,omitempty"`
	OriginatorInfos      []OriginatorInfosRequest `json:"originatorInfos,omitempty"`
}

type UpdateBankTransferStatusRequest struct {
	ExternalID string `json:"externalId"`
	Status     string `json:"status"`
}

type BTBeneficiaryRequest struct {
	BeneficiaryBankCode          string `json:"beneficiaryBankCode" validate:"required"`
	BeneficiaryAccountNo         string `json:"beneficiaryAccountNo" validate:"required"`
	BeneficiaryAccountName       string `json:"beneficiaryAccountName" validate:"required"`
	BeneficiaryEmail             string `json:"beneficiaryEmail,omitempty"`
	BeneficiaryCustomerResidence string `json:"beneficiaryCustomerResidence,omitempty"`
	BeneficiaryAddress           string `json:"beneficiaryAddress,omitempty"`
	BeneficiaryCustomerType      string `json:"beneficiaryCustomerType,omitempty"`
	BeneficiaryCitizenStatus     string `json:"beneficiaryCitizenStatus,omitempty"`
	BeneficiaryBICCode           string `json:"beneficiaryBICCode,omitempty"`
}

type BankTransferHeaderRequest struct {
	ExternalId string `json:"X-EXTERNAL-ID"`
	MerchantId string `json:"X-MERCHANT-ID"`
}

type OriginatorInfosRequest struct {
	OriginatorCustomerNo   string `json:"originatorCustomerNo,omitempty"`
	OriginatorCustomerName string `json:"originatorCustomerName,omitempty"`
	OriginatorBankCode     string `json:"originatorBankCode,omitempty"`
}
