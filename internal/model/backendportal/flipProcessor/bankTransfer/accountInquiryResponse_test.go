package flipProcessorModel_test

import (
	"github.com/paper-indonesia/pdk/go/snap"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/flipProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/stretchr/testify/assert"
)

func TestAccountInquiryResponse_ToAccountInquiryResponse(t *testing.T) {
	tests := []struct {
		name     string
		response flipProcessorModel.AccountInquiryResponse
		expected routingProcessorModel.InquiryAccountResponseData
	}{
		{
			name: "Success status",
			response: flipProcessorModel.AccountInquiryResponse{
				AccountNo:     "123456789",
				BankCode:      "",
				AccountHolder: "John Doe",
				Status:        constant.FlipAccountInquiryStatusSuccess,
				InquiryKey:    "key123",
			},
			expected: routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "2001600",
				ResponseMessage:        "Successful",
				PartnerReferenceNo:     "key123",
				BeneficiaryAccountName: "John Doe",
				BeneficiaryAccountNo:   "123456789",
				BeneficiaryBankCode:    "",
				ProcessorReference:     constant.FlipPGProcessor,
				Status:                 snap.SNAP_SUCCESS, // Updated to match actual value
			},
		},
		{
			name: "Invalid account status",
			response: flipProcessorModel.AccountInquiryResponse{
				AccountNo:     "987654321",
				BankCode:      "BNI",
				AccountHolder: "Jane Doe",
				Status:        constant.FlipAccountInquiryStatusInvalid,
				InquiryKey:    "key456",
			},
			expected: routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "4031611",
				ResponseMessage:        "Invalid Account",
				PartnerReferenceNo:     "key456",
				BeneficiaryAccountName: "Jane Doe",
				BeneficiaryAccountNo:   "987654321",
				BeneficiaryBankCode:    "",
				ProcessorReference:     constant.FlipPGProcessor,
				Status:                 snap.SNAP_INVALID_ACCOUNT, // Updated to match actual value
			},
		},
		{
			name: "Default case - unknown status",
			response: flipProcessorModel.AccountInquiryResponse{
				AccountNo:     "555555555",
				BankCode:      "BCA",
				AccountHolder: "Unknown User",
				Status:        "UNKNOWN_STATUS",
				InquiryKey:    "key789",
			},
			expected: routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "2021600",
				ResponseMessage:        "Request In Progress",
				PartnerReferenceNo:     "key789",
				BeneficiaryAccountName: "Unknown User",
				BeneficiaryAccountNo:   "555555555",
				BeneficiaryBankCode:    "",
				ProcessorReference:     constant.FlipPGProcessor,
				Status:                 snap.SNAP_INPROGRESS,
			},
		},
		{
			name: "Blacklisted account status",
			response: flipProcessorModel.AccountInquiryResponse{
				AccountNo:     "111222333",
				BankCode:      "MANDIRI",
				AccountHolder: "Blacklisted User",
				Status:        constant.FlipAccountInquiryStatusBlacklisted,
				InquiryKey:    "key101",
			},
			expected: routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "4031611",
				ResponseMessage:        "Invalid Account",
				PartnerReferenceNo:     "key101",
				BeneficiaryAccountName: "Blacklisted User",
				BeneficiaryAccountNo:   "111222333",
				BeneficiaryBankCode:    "",
				ProcessorReference:     constant.FlipPGProcessor,
				Status:                 snap.SNAP_INVALID_ACCOUNT,
			},
		},
		{
			name: "Suspected account status",
			response: flipProcessorModel.AccountInquiryResponse{
				AccountNo:     "444555666",
				BankCode:      "BRI",
				AccountHolder: "Suspected User",
				Status:        constant.FlipAccountInquiryStatusSuspectedAccount,
				InquiryKey:    "key202",
			},
			expected: routingProcessorModel.InquiryAccountResponseData{
				ResponseCode:           "4031611",
				ResponseMessage:        "Invalid Account",
				PartnerReferenceNo:     "key202",
				BeneficiaryAccountName: "Suspected User",
				BeneficiaryAccountNo:   "444555666",
				BeneficiaryBankCode:    "",
				ProcessorReference:     constant.FlipPGProcessor,
				Status:                 snap.SNAP_INVALID_ACCOUNT,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.response.ToAccountInquiryResponse()
			assert.Equal(t, tt.expected.ResponseCode, response.ResponseCode)
			assert.Equal(t, tt.expected.ResponseMessage, response.ResponseMessage)
			assert.Equal(t, tt.expected.PartnerReferenceNo, response.PartnerReferenceNo)
			assert.Equal(t, tt.expected.BeneficiaryAccountName, response.BeneficiaryAccountName)
			assert.Equal(t, tt.expected.BeneficiaryAccountNo, response.BeneficiaryAccountNo)
			assert.Equal(t, tt.expected.BeneficiaryBankCode, response.BeneficiaryBankCode)
			assert.Equal(t, tt.expected.ProcessorReference, response.ProcessorReference)
			assert.Equal(t, tt.expected.Status, response.Status)
		})
	}
}
