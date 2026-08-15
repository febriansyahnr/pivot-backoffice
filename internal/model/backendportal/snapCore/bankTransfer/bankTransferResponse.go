package banktransfer

import (
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

type BankTransferResponse struct {
	Data    BankTransferResponseData `json:"data"`
	Code    string                   `json:"code"`
	Message string                   `json:"message"`
	Error   interface{}              `json:"error,omitempty"`
}

type BankTransferResponseData struct {
	ResponseCode           string             `json:"responseCode"`
	ResponseMessage        string             `json:"responseMessage"`
	UUID                   string             `json:"uuid"`
	PartnerReferenceNo     string             `json:"partnerReferenceNo,omitempty"`
	BankReferenceNo        string             `json:"bankReferenceNo,omitempty"`
	BankProcessor          string             `json:"bankProcessor,omitempty"`
	Amount                 commonModel.Amount `json:"amount,omitempty"`
	BeneficiaryAccountNo   string             `json:"beneficiaryAccountNo,omitempty"`
	BeneficiaryBankCode    string             `json:"beneficiaryBankCode,omitempty"`
	BeneficiaryAccountName string             `json:"beneficiaryAccountName,omitempty"`
	SourceAccountNo        string             `json:"sourceAccountNo,omitempty"`
	Status                 string             `json:"status"`
	TransferType           string             `json:"transferType"`
	ExternalID             string             `json:"externalId,omitempty"`
	TransactionDate        time.Time          `json:"transactionDate,omitempty"`
	Remark                 string             `json:"remark,omitempty"`
	AdditionalInfo         map[string]any     `json:"additionalInfo,omitempty"`
}

func (d *BankTransferResponseData) MappingAccountTransactionErrStatus() (status, reasonType, reasonDesc string) {
	status = constant.StatusFailed
	reasonType, reasonDesc = constant.ReasonTypeOtherReason, ""

	// Set status to PENDING due to SnapCoreResponseCodeInsufficientFund
	if d.Status == constant.SnapCoreBankTransferStatusPending {
		status = constant.StatusPending
		reasonDesc = d.ResponseMessage

	} else if d.Status == constant.SnapCoreBankTransferStatusFailed &&
		util.IsPatternMatch(constant.SnapCoreResponseCodeInsufficientFundPattern, d.ResponseCode) {

		reasonType = constant.ReasonTypeInsufficientEscrowFund
		reasonDesc = d.ResponseMessage
		status = constant.StatusPending

	} else if util.IsPatternMatch(constant.SnapCoreResponseCodeInactiveAccountPattern, d.ResponseCode) {
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseInactiveAccountMessage

	} else if util.IsPatternMatch(constant.SnapCoreResponseCodeDormantAccountPattern, d.ResponseCode) {
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseDormantAccountMessage

	} else if util.IsPatternMatch(constant.SnapCoreResponseCodeInvalidAccountPattern, d.ResponseCode) {
		reasonType = constant.ReasonTypeBeneficiaryAccountReason
		reasonDesc = constant.SnapCoreResponseInvalidAccountMessage

	} else if d.ResponseMessage != "" {
		reasonDesc = d.ResponseMessage
	}

	return
}

func (d *BankTransferResponseData) MappingInquiryTransactionStatus() (status, reasonType, reasonDesc string) {

	status = constant.StatusPending

	switch d.Status {
	case constant.SnapCoreBankTransferStatusSuccess:

		status = constant.StatusSuccess

	case constant.SnapCoreBankTransferStatusFailed:

		status = constant.StatusFailed
		reasonType, reasonDesc = constant.ReasonTypeOtherReason, d.ResponseMessage
	}

	// Set status to PENDING due to SnapCoreResponseCodeInsufficientFund
	if d.Status == constant.SnapCoreBankTransferStatusFailed && util.IsPatternMatch(constant.SnapCoreResponseCodeInsufficientFundPattern, d.ResponseCode) {

		status = constant.StatusPending
		reasonType, reasonDesc = constant.ReasonTypeInsufficientEscrowFund, d.ResponseMessage
	}
	return
}

func (d *BankTransferResponseData) GetReconReferenceNo() string {
	switch d.BankProcessor {
	case "BNI", "DANA":
		return d.PartnerReferenceNo
	case "BRI":
		if d.TransferType != "INTERBANK-BIFAST" {
			return d.PartnerReferenceNo
		}
	case "DBS":
		// For DBS use snap_core.bank_transfer.additional_info.txnResponse.customerReference
		if val, ok := d.AdditionalInfo["txnResponse"].(map[string]any); ok {
			if customerReference, ok := val["customerReference"].(string); ok {
				return customerReference
			}
		}
	case "MANDIRI_CENTRAL":
		// For MANDIRI_CENTRAL to use snap_core.bank_transfer.additional_info.partnerExternalId
		if val, ok := d.AdditionalInfo["partnerExternalId"].(string); ok {
			return val
		}
	case "BCA":
		// BCA need mix between amount + remark + beneficiary account name
		return fmt.Sprintf("%s %s %s", d.Amount.Value, d.Remark, d.BeneficiaryAccountName)
	}

	// Unspecified bank processor will use bank reference no
	return d.BankReferenceNo
}
