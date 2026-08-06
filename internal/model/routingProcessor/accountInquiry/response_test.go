package routingProcessorModel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInquiryAccountResponseData_ToSnapCoreResponseData(t *testing.T) {
	// Test cases
	testCases := []struct {
		name     string
		input    InquiryAccountResponseData
		expected InquiryAccountResponseData
	}{
		{
			name: "Complete data conversion",
			input: InquiryAccountResponseData{
				ResponseCode:           "200xx200",
				ResponseMessage:        "Success",
				PartnerReferenceNo:     "BT-120",
				BeneficiaryAccountName: "John Doe",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryBankCode:    "013",
				BeneficiaryBankName:    "PERMATA",
				IsVirtualAccount:       true,
				ProcessorReference:     "PROC-REF-123", // This field should not be copied
				Status:                 "COMPLETED",    // This field should not be copied
			},
			expected: InquiryAccountResponseData{
				ResponseCode:           "200xx200",
				ResponseMessage:        "Success",
				PartnerReferenceNo:     "BT-120",
				BeneficiaryAccountName: "John Doe",
				BeneficiaryAccountNo:   "1234567890",
				BeneficiaryBankCode:    "013",
				BeneficiaryBankName:    "PERMATA",
				IsVirtualAccount:       true,
				// ProcessorReference and Status are not expected in the output
			},
		},
		{
			name: "Partial data conversion",
			input: InquiryAccountResponseData{
				ResponseCode:           "400xx400",
				ResponseMessage:        "Failed",
				PartnerReferenceNo:     "BT-121",
				BeneficiaryAccountName: "",
				BeneficiaryAccountNo:   "",
				BeneficiaryBankCode:    "",
				BeneficiaryBankName:    "",
				IsVirtualAccount:       false,
				ProcessorReference:     "PROC-REF-124",
				Status:                 "FAILED",
			},
			expected: InquiryAccountResponseData{
				ResponseCode:           "400xx400",
				ResponseMessage:        "Failed",
				PartnerReferenceNo:     "BT-121",
				BeneficiaryAccountName: "",
				BeneficiaryAccountNo:   "",
				BeneficiaryBankCode:    "",
				BeneficiaryBankName:    "",
				IsVirtualAccount:       false,
				// ProcessorReference and Status are not expected in the output
			},
		},
		{
			name: "Empty data conversion",
			input: InquiryAccountResponseData{
				ResponseCode:           "",
				ResponseMessage:        "",
				PartnerReferenceNo:     "",
				BeneficiaryAccountName: "",
				BeneficiaryAccountNo:   "",
				BeneficiaryBankCode:    "",
				BeneficiaryBankName:    "",
				IsVirtualAccount:       true,
				ProcessorReference:     "",
				Status:                 "",
			},
			expected: InquiryAccountResponseData{
				ResponseCode:           "",
				ResponseMessage:        "",
				PartnerReferenceNo:     "",
				BeneficiaryAccountName: "",
				BeneficiaryAccountNo:   "",
				BeneficiaryBankCode:    "",
				BeneficiaryBankName:    "",
				IsVirtualAccount:       true,
				// ProcessorReference and Status are not expected in the output
			},
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Call the method
			result := tc.input.ToSnapCoreResponseData()

			// Verify the result
			assert.Equal(t, tc.expected.ResponseCode, result.ResponseCode, "ResponseCode should match")
			assert.Equal(t, tc.expected.ResponseMessage, result.ResponseMessage, "ResponseMessage should match")
			assert.Equal(t, tc.expected.PartnerReferenceNo, result.PartnerReferenceNo, "PartnerReferenceNo should match")
			assert.Equal(t, tc.expected.BeneficiaryAccountName, result.BeneficiaryAccountName, "BeneficiaryAccountName should match")
			assert.Equal(t, tc.expected.BeneficiaryAccountNo, result.BeneficiaryAccountNo, "BeneficiaryAccountNo should match")
			assert.Equal(t, tc.expected.BeneficiaryBankCode, result.BeneficiaryBankCode, "BeneficiaryBankCode should match")
			assert.Equal(t, tc.expected.BeneficiaryBankName, result.BeneficiaryBankName, "BeneficiaryBankName should match")
			assert.Equal(t, tc.expected.IsVirtualAccount, result.IsVirtualAccount, "IsVirtualAccount should match")
		})
	}
}
