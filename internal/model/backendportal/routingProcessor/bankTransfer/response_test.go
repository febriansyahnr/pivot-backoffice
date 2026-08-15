package routingProcessorModel_test

import (
	"testing"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"

	"github.com/stretchr/testify/assert"
)

func TestBankTransferResponseData(t *testing.T) {
	input := BankTransferResponseData{
		ResponseCode:       "ResponseCode",
		ResponseMessage:    "ResponseMessage",
		UUID:               "UUID",
		PartnerReferenceNo: "PartnerReferenceNo",
		BankReferenceNo:    "BankReferenceNo",
		BankProcessor:      "BankProcessor",
		Amount: commonModel.Amount{
			Currency: "IDR", Value: "10000.00",
		},
		BeneficiaryAccountNo: "BeneficiaryAccountNo",
		BeneficiaryBankCode:  "BeneficiaryBankCode",
		SourceAccountNo:      "SourceAccountNo",
		Status:               "Status",
		TransferType:         "TransferType",
		ExternalID:           "ExternalID",
	}
	wantResult := snapCoreModel.BankTransferResponseData{
		ResponseCode:       "ResponseCode",
		ResponseMessage:    "ResponseMessage",
		UUID:               "UUID",
		PartnerReferenceNo: "PartnerReferenceNo",
		BankReferenceNo:    "BankReferenceNo",
		BankProcessor:      "BankProcessor",
		Amount: commonModel.Amount{
			Currency: "IDR", Value: "10000.00",
		},
		BeneficiaryAccountNo: "BeneficiaryAccountNo",
		BeneficiaryBankCode:  "BeneficiaryBankCode",
		SourceAccountNo:      "SourceAccountNo",
		Status:               "Status",
		TransferType:         "TransferType",
		ExternalID:           "ExternalID",
	}
	assert.Equal(t, wantResult, input.ToSnapBankTransferResponseData())
}
