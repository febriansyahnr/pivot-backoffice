package disbursementModel_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/fee"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisbursementForReversal(t *testing.T) {
	data := &DisbursementForReversal{
		Fee: TransactionMetadataForReversal{
			Status: c.StatusPending,
			Metadata: feeModel.FeeMetadataObject{
				DeductionType: c.MerchantFeeDeductionTypeAutomated,
			},
		},
	}

	assert.False(t, data.IsFeeStatus(c.StatusSuccess))
	assert.True(t, data.IsFeeStatus(c.StatusPending))
	assert.False(t, data.IsFeeDeductionType(c.MerchantFeeDeductionTypeDirect))
	assert.True(t, data.IsFeeDeductionType(c.MerchantFeeDeductionTypeAutomated))
}

func TestValidateDisbursementType(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Single Disbursement",
			input:   c.DisbursementTypeSingle,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Bulk Disbursement",
			input:   c.DisbursementTypeBulk,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Single Disbursement lowercase",
			input:   strings.ToLower(c.DisbursementTypeSingle),
			wantErr: false,
		},
		{
			name:    "SUCCESS: Bulk Disbursement lowercase",
			input:   strings.ToLower(c.DisbursementTypeBulk),
			wantErr: false,
		},
		{
			name:    "SUCCESS: Single Disbursement mixed case",
			input:   "Single",
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Disbursement",
			input:   "MIXED",
			wantErr: true,
		},
		{
			name:    "ERROR: Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "ERROR: Special characters",
			input:   "SINGLE!@#",
			wantErr: true,
		},
		{
			name:    "ERROR: Numeric input",
			input:   "123",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateDisbursementType(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, c.ErrInvalidDisbursementType, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateApprovalStatus(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Pending Status",
			input:   c.DisbursementStatusPending,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Waiting Status",
			input:   c.DisbursementStatusWaiting,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Approved Status",
			input:   c.DisbursementStatusApproved,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Rejected Status",
			input:   c.DisbursementStatusRejected,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Pending Status lowercase",
			input:   strings.ToLower(c.DisbursementStatusPending),
			wantErr: false,
		},
		{
			name:    "SUCCESS: Waiting Status mixed case",
			input:   "Waiting",
			wantErr: false,
		},
		{
			name:    "ERROR: Unknown Status",
			input:   "UNKNOWN",
			wantErr: true,
		},
		{
			name:    "ERROR: Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "ERROR: Special characters",
			input:   "PENDING!@#",
			wantErr: true,
		},
		{
			name:    "ERROR: Numeric status",
			input:   "123",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateApprovalStatus(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, c.ErrInvalidDisbursementApprovalStatus, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestPaymentStatus(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Pending Status",
			input:   c.StatusPending,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Success Status",
			input:   c.StatusSuccess,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Failed Status",
			input:   c.StatusFailed,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Cancelled Status",
			input:   c.SettlementStatusCancelled,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Pending Status lowercase",
			input:   strings.ToLower(c.StatusPending),
			wantErr: false,
		},
		{
			name:    "SUCCESS: Success Status mixed case",
			input:   "Success",
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Status",
			input:   "MIXED",
			wantErr: true,
		},
		{
			name:    "ERROR: Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "ERROR: Special characters",
			input:   "SUCCESS!@#",
			wantErr: true,
		},
		{
			name:    "ERROR: Numeric status",
			input:   "999",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateTransactionStatus(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, c.ErrInvalidDisbursementPaymentStatus, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestValidateSortColumn(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: UpdatedAt Column",
			input:   "updatedAt",
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Sort Column",
			input:   "MIXED",
			wantErr: true,
		},
		{
			name:    "ERROR: Empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "SUCCESS: createdAt column",
			input:   "createdAt",
			wantErr: false,
		},
		{
			name:    "ERROR: UpdatedAt with different case",
			input:   "UpdatedAt",
			wantErr: true,
		},
		{
			name:    "ERROR: Special characters",
			input:   "updatedAt!@#",
			wantErr: true,
		},
		{
			name:    "ERROR: Numeric column",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "ERROR: SQL injection attempt",
			input:   "updatedAt; DROP TABLE",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSortColumn(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, c.ErrInvalidDisbursementListSortColumn, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestTransactionConfigMarshalBinary(t *testing.T) {
	// Create a TransactionConfig instance
	config := TransactionConfig{
		MinAmount: 10000.0,
		MaxAmount: 100000.0,
	}

	// Test MarshalBinary
	data, err := config.MarshalBinary()
	require.NoError(t, err)
	require.NotNil(t, data)

	// Verify the marshaled data
	var unmarshaledConfig map[string]interface{}
	err = json.Unmarshal(data, &unmarshaledConfig)
	require.NoError(t, err)

	assert.Equal(t, 10000.0, unmarshaledConfig["minAmount"])
	assert.Equal(t, 100000.0, unmarshaledConfig["maxAmount"])
}

func TestTransactionConfigUnmarshalBinary(t *testing.T) {
	// Create test data
	jsonData := []byte(`{"minAmount":10000,"maxAmount":100000}`)

	// Test UnmarshalBinary
	var config TransactionConfig
	err := config.UnmarshalBinary(jsonData)
	require.NoError(t, err)

	// Verify the unmarshaled data
	assert.Equal(t, 10000.0, config.MinAmount)
	assert.Equal(t, 100000.0, config.MaxAmount)

	// Test with invalid JSON
	invalidJSON := []byte(`{"minAmount":10000,"maxAmount":invalid}`)
	err = config.UnmarshalBinary(invalidJSON)
	assert.Error(t, err, "Should return an error for invalid JSON")
}

func TestApprovalResultErr_Error(t *testing.T) {
	tests := []struct {
		name              string
		approvalResultErr *ApprovalResultErr
		expectedErrorMsg  string
	}{
		{
			name: "empty_beneficiary_limit_exceeded",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "single_beneficiary_limit_exceeded",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "1234567890",
						Amount:    1000000.0,
						Error:     c.ErrInvalidDisbursementApprovalStatus,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "multiple_beneficiary_limit_exceeded",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "1234567890",
						Amount:    1000000.0,
						Error:     c.ErrInvalidDisbursementApprovalStatus,
					},
					{
						AccountNo: "0987654321",
						Amount:    2000000.0,
						Error:     c.ErrInvalidDisbursementType,
					},
					{
						AccountNo: "1111222233",
						Amount:    500000.0,
						Error:     nil,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "beneficiary_with_zero_amount",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "0000000000",
						Amount:    0.0,
						Error:     c.ErrInvalidDisbursementPaymentStatus,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "beneficiary_with_negative_amount",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "9999999999",
						Amount:    -1000.0,
						Error:     c.ErrInvalidDisbursementListSortColumn,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "beneficiary_with_empty_account_no",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "",
						Amount:    500000.0,
						Error:     c.ErrInvalidDisbursementType,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "beneficiary_with_special_characters_in_account",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "1234-5678-90AB",
						Amount:    750000.0,
						Error:     nil,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
		{
			name: "beneficiary_with_very_large_amount",
			approvalResultErr: &ApprovalResultErr{
				BeneficiaryLimitExceeded: []ApprovalValidation{
					{
						AccountNo: "9876543210",
						Amount:    999999999999.99,
						Error:     c.ErrInvalidDisbursementApprovalStatus,
					},
				},
			},
			expectedErrorMsg: c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.approvalResultErr.Error()
			assert.Equal(t, tt.expectedErrorMsg, result)
			assert.NotEmpty(t, result)
			assert.IsType(t, "", result)
		})
	}
}

// Test to ensure the Error() method always returns the same constant regardless of the struct content
func TestApprovalResultErr_Error_Consistency(t *testing.T) {
	// Test various combinations to ensure consistency
	variations := []*ApprovalResultErr{
		{BeneficiaryLimitExceeded: nil},
		{BeneficiaryLimitExceeded: []ApprovalValidation{}},
		{BeneficiaryLimitExceeded: []ApprovalValidation{{AccountNo: "test", Amount: 100, Error: nil}}},
		{BeneficiaryLimitExceeded: make([]ApprovalValidation, 100)}, // Large slice
	}

	expectedMsg := c.ErrMsgPayoutDeclinedDueToBeneficiaryLimitRestrictions

	for i, variation := range variations {
		t.Run(fmt.Sprintf("variation_%d", i), func(t *testing.T) {
			result := variation.Error()
			assert.Equal(t, expectedMsg, result, "All variations should return the same error message")
		})
	}
}
