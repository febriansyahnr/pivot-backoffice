package flipProcessorModel_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/routingProcessor/accountInquiry"
	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
	"github.com/stretchr/testify/assert"
)

func TestToAccountInquiryResponse(t *testing.T) {
	tests := []struct {
		name     string
		request  flipProcessorModel.AccountInquiryRequest
		expected routingProcessorModel.InquiryAccountResponseData
	}{
		{
			name: "Success status",
			request: flipProcessorModel.AccountInquiryRequest{
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
				Status:                 constant.FlipAccountInquiryStatusSuccess,
			},
		},
		{
			name: "Invalid account status",
			request: flipProcessorModel.AccountInquiryRequest{
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
				Status:                 constant.FlipAccountInquiryStatusInvalid,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := tt.request.ToAccountInquiryResponse()
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
