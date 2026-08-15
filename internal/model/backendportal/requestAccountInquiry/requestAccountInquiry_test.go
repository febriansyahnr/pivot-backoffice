package requestAccountInquiries

import (
	"encoding/json"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankAccount"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestNewDetailStatusRequestInquiry(t *testing.T) {
	accountName := "John Doe"
	_, expectedWarningStatus := util.SimilarityCheck("John Doe", "Sdr John Doe", "", "")
	_, expectedEmptyAccount := util.SimilarityCheck("", "John Doe", "", "")

	testCases := []struct {
		name                  string
		status                string
		accountName           string
		bankRecord            string
		expected              string
		processorResponseCode string
	}{
		{
			name:                  "Invalid status",
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           "John Doe",
			expected:              constant.ReqInquiryDetailStatusInvalid,
			processorResponseCode: "",
		},
		{
			name:                  "Pending status",
			status:                constant.RequestAccountInquiryStatusPending,
			accountName:           "John Doe",
			expected:              constant.ReqInquiryDetailStatusPending,
			processorResponseCode: "",
		},
		{
			name:                  "Warning status",
			status:                constant.RequestAccountInquiryStatusWarning,
			bankRecord:            "Sdr John Doe",
			accountName:           "John Doe",
			expected:              expectedWarningStatus,
			processorResponseCode: "",
		},
		{
			name:                  "Empty account name with warning status",
			status:                constant.RequestAccountInquiryStatusWarning,
			accountName:           "",
			bankRecord:            "John Doe",
			expected:              expectedEmptyAccount,
			processorResponseCode: "",
		},
		{
			name:                  "Unknown status",
			status:                "UNKNOWN",
			accountName:           "John Doe",
			expected:              "",
			processorResponseCode: "",
		},
		{
			name:                  "Empty status",
			status:                "",
			accountName:           "John Doe",
			expected:              "",
			processorResponseCode: "",
		},
		{
			name:                  "Invalid Account", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusInvalid,
			processorResponseCode: "403xx011", // NOSONAR
		},

		{
			name:                  "Dormant Account", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusDormant,
			processorResponseCode: "403xx09", // NOSONAR
		},
		{
			name:                  "Inactive Account", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusInactive,
			processorResponseCode: "403xx18", // NOSONAR
		},
		{
			name:                  "Suspected Fraud", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusSuspectedFraud,
			processorResponseCode: "403xx03", // NOSONAR
		},
		{
			name:                  "Activity Count Limit Exceeded", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusLimitExceeded,
			processorResponseCode: "403xx04", // NOSONAR
		},
		{
			name:                  "Do Not Honor", // NOSONAR
			status:                constant.RequestAccountInquiryStatusInvalid,
			accountName:           accountName,
			expected:              constant.ReqInquiryDetailStatusDoNotHonor,
			processorResponseCode: "403xx05", // NOSONAR
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, result := NewDetailStatusRequestInquiry(tc.status, tc.accountName, tc.bankRecord, "", tc.processorResponseCode)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSetMetadataNullJSONText(t *testing.T) {
	testCases := []struct {
		name     string
		metadata Metadata
	}{
		{
			name:     "Empty metadata",
			metadata: Metadata{},
		},
		{
			name: "With detail status only",
			metadata: Metadata{
				DetailStatus: "Account number not found.",
			},
		},
		{
			name: "With SnapCore response",
			metadata: Metadata{
				DetailStatus: "Account number not found.",
				SnapCoreResponse: &snapCoreModel.InquiryAccountResponseData{
					ResponseCode:           "200xx200",
					ResponseMessage:        "Success",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "013",
					BeneficiaryBankName:    "PERMATA",
				},
			},
		},
		{
			name: "With OnBehalf data",
			metadata: Metadata{
				DetailStatus: "Account number not found.",
				OnBehalf: &merchant.OnBehalfObject{
					ParentMerchantId: "merchant-123",
				},
			},
		},
		{
			name: "With fee details",
			metadata: Metadata{
				DetailStatus: "Account number not found.",
				FeeDetail: &feeModel.FeeMetadataObject{
					Amount: 5000,
				},
			},
		},
		{
			name: "With all fields",
			metadata: Metadata{
				DetailStatus: "Account number not found.",
				SnapCoreResponse: &snapCoreModel.InquiryAccountResponseData{
					ResponseCode:           "200xx200",
					ResponseMessage:        "Success",
					PartnerReferenceNo:     "BT-120",
					BeneficiaryAccountName: "John Doe",
					BeneficiaryAccountNo:   "1234567890",
					BeneficiaryBankCode:    "013",
					BeneficiaryBankName:    "PERMATA",
				},
				OnBehalf: &merchant.OnBehalfObject{
					ParentMerchantId: "merchant-123",
				},
				FeeOnBehalf: &feeModel.TrxFeeOnBehalfMetadata{
					Amount: 5000,
				},
				FeeDetail: &feeModel.FeeMetadataObject{
					Amount: 5000,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a RequestAccountInquiryWithMaster with the test metadata
			inquiry := RequestAccountInquiryWithMaster{}
			inquiry.MetadataObj = tc.metadata

			// Call the method
			inquiry.SetMetadataNullJSONText()

			// Verify the result
			assert.True(t, inquiry.Metadata.Valid, "Metadata should be valid")

			// Unmarshal the JSON text back to a Metadata object
			var resultMetadata Metadata
			err := json.Unmarshal(inquiry.Metadata.JSONText, &resultMetadata)
			assert.NoError(t, err, "Should unmarshal without error")

			// Compare the original and unmarshaled metadata
			// For DetailStatus
			assert.Equal(t, tc.metadata.DetailStatus, resultMetadata.DetailStatus, "DetailStatus should match")

			// For SnapCoreResponse
			if tc.metadata.SnapCoreResponse != nil {
				assert.NotNil(t, resultMetadata.SnapCoreResponse, "SnapCoreResponse should not be nil")
				assert.Equal(t, tc.metadata.SnapCoreResponse.ResponseCode, resultMetadata.SnapCoreResponse.ResponseCode, "ResponseCode should match")
				assert.Equal(t, tc.metadata.SnapCoreResponse.BeneficiaryAccountName, resultMetadata.SnapCoreResponse.BeneficiaryAccountName, "BeneficiaryAccountName should match")
				assert.Equal(t, tc.metadata.SnapCoreResponse.BeneficiaryAccountNo, resultMetadata.SnapCoreResponse.BeneficiaryAccountNo, "BeneficiaryAccountNo should match")
			} else {
				assert.Nil(t, resultMetadata.SnapCoreResponse, "SnapCoreResponse should be nil")
			}

			// For OnBehalf
			if tc.metadata.OnBehalf != nil {
				assert.NotNil(t, resultMetadata.OnBehalf, "OnBehalf should not be nil")
				assert.Equal(t, tc.metadata.OnBehalf.ParentMerchantId, resultMetadata.OnBehalf.ParentMerchantId, "ParentMerchantId should match")
			} else {
				assert.Nil(t, resultMetadata.OnBehalf, "OnBehalf should be nil")
			}

			// For FeeOnBehalf
			if tc.metadata.FeeOnBehalf != nil {
				assert.NotNil(t, resultMetadata.FeeOnBehalf, "FeeOnBehalf should not be nil")
				assert.Equal(t, tc.metadata.FeeOnBehalf.Amount, resultMetadata.FeeOnBehalf.Amount, "Amount should match")
			} else {
				assert.Nil(t, resultMetadata.FeeOnBehalf, "FeeOnBehalf should be nil")
			}

			// For FeeDetail
			if tc.metadata.FeeDetail != nil {
				assert.NotNil(t, resultMetadata.FeeDetail, "FeeDetail should not be nil")
				assert.Equal(t, tc.metadata.FeeDetail.Amount, resultMetadata.FeeDetail.Amount, "Amount should match")
			} else {
				assert.Nil(t, resultMetadata.FeeDetail, "FeeDetail should be nil")
			}

			// Alternative approach: compare the JSON strings directly
			expectedJSON, _ := json.Marshal(tc.metadata)
			assert.JSONEq(t, string(expectedJSON), string(inquiry.Metadata.JSONText), "JSON representation should match")
		})
	}
}
