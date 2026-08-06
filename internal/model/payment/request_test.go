package paymentModel

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"

	"github.com/stretchr/testify/assert"
)

func TestQueryQrMpStaticRequestValidation(t *testing.T) {
	testCase := []struct {
		name     string
		req      QueryQrMpmStaticRequest
		expected QueryQrMpmStaticRequest
	}{
		{
			name: "default validation",
			req: QueryQrMpmStaticRequest{
				ReferenceId:  "123",
				FromDateTime: "invalid",
				ToDateTime:   "2024-07-18T17:11:43+07:00",
			},
			expected: QueryQrMpmStaticRequest{
				ReferenceId: "123",
				PageNumber:  1,
				PageSize:    20,
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			tc.req.Validate()
			assert.Equal(t, tc.expected.PageNumber, tc.req.PageNumber)
			assert.Equal(t, tc.expected.PageSize, tc.req.PageSize)
		})
	}

}

func TestFilterPaymentHistoryOptionValidate(t *testing.T) {
	testCases := []struct {
		name    string
		input   FilterPaymentHistoryOption
		wantErr bool
	}{
		{
			name:  "SUCCESS: Valid Request",
			input: FilterPaymentHistoryOption{},
		},
		{
			name: "SUCCESS: Valid Payment Method",
			input: FilterPaymentHistoryOption{
				PaymentMethod: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			},
		},
		{
			name: "SUCCESS: Valid Payment Status",
			input: FilterPaymentHistoryOption{
				Status: paymentConstant.PAYMENT_STATUS_PENDING,
			},
		},
		{
			name: "SUCCESS: Valid Payment Sort Column",
			input: FilterPaymentHistoryOption{
				SortBy: "createdAt",
			},
		},
		{
			name: "ERROR: Invalid Payment Method",
			input: FilterPaymentHistoryOption{
				PaymentMethod: "STONE",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Payment Status",
			input: FilterPaymentHistoryOption{
				Status: "UNKNOWN",
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Disbursement Sort",
			input: FilterPaymentHistoryOption{
				SortBy: "updatedAts",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestPaymentDownloadHistoryRequest(t *testing.T) {
	request := &PaymentDownloadHistoryRequest{
		MerchantId: "e69fd166-76dc-4b0a-8468-246792f2c780",
		StartDate:  "2024-11-01",
		EndDate:    "2024-11-17",
	}
	result := request.ToFilterPaymentHistoryOption()

	t.Run("TestToFilterPaymentHistoryOption", func(t *testing.T) {
		want := FilterPaymentHistoryOption{
			MerchantID: "e69fd166-76dc-4b0a-8468-246792f2c780",
			Sort:       "ASC",
			SortBy:     "createdAt",
			Page:       1, PerPage: 1_048_576,
		}
		want.StartDate, _ = time.ParseInLocation(time.DateTime, request.StartDate+" 00:00:00", loc)
		want.EndDate, _ = time.ParseInLocation(time.DateTime, request.EndDate+" 23:59:59", loc)

		assert.Equal(t, want, result)
	})

	t.Run("TestHashFilterKey", func(t *testing.T) {
		assert.Equal(t, "682448b94f95589c9ae13e6d2872d527f8e30c8c642df94800129d43f667bdca", request.HashFilterKey(result.EndDate))
	})
}

func TestAutoSplitPaymentSummaryGetFinalStatus(t *testing.T) {
	const numberOfCharges = 3
	tests := []struct {
		name     string
		input    AutoSplitPaymentSummary
		expected string
	}{
		{
			name: "no charges returns empty",
			input: AutoSplitPaymentSummary{
				NumberOfCharges: 0,
			},
			expected: "PROCESSING",
		},
		{
			name: "all failed",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:       numberOfCharges,
				NumberOfFailedCharges: numberOfCharges,
			},
			expected: constant.AutoSplitPaymentStatusFailed,
		},
		{
			name: "all expired",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:        numberOfCharges,
				NumberOfExpiredCharges: numberOfCharges,
			},
			expected: constant.AutoSplitPaymentStatusCancelled,
		},
		{
			name: "all success",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:           numberOfCharges,
				NumberOfSuccessfulCharges: numberOfCharges,
			},
			expected: constant.AutoSplitPaymentStatusSuccess,
		},
		{
			name: "partial success (mix of all states)",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:           5,
				NumberOfSuccessfulCharges: 2,
				NumberOfFailedCharges:     2,
				NumberOfExpiredCharges:    1,
			},
			expected: constant.AutoSplitPaymentStatusPartialSuccess,
		},
		{
			name: "processing when not all accounted",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:           5,
				NumberOfSuccessfulCharges: 2,
				NumberOfFailedCharges:     1,
				NumberOfExpiredCharges:    1,
				// 1 remaining = in progress
			},
			expected: constant.AutoSplitPaymentStatusProcessing,
		},
		{
			name: "failed takes precedence over partial",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:           numberOfCharges,
				NumberOfFailedCharges:     numberOfCharges,
				NumberOfSuccessfulCharges: 1, // inconsistent but tests precedence
			},
			expected: constant.AutoSplitPaymentStatusFailed,
		},
		{
			name: "expired takes precedence over partial",
			input: AutoSplitPaymentSummary{
				NumberOfCharges:        numberOfCharges,
				NumberOfExpiredCharges: numberOfCharges,
				NumberOfFailedCharges:  1,
			},
			expected: constant.AutoSplitPaymentStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.GetFinalStatus()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoSplitPaymentSummaryUpdateRecordByParentStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		payload  *AutoSplitPaymentSummary
		expected AutoSplitPaymentSummary
	}{
		{
			name:   "PAID - sets NumberOfSuccessfulCharges to 1",
			status: constant.UnifiedPaymentSessionStatusPaid,
			expected: AutoSplitPaymentSummary{
				NumberOfSuccessfulCharges: 1,
			},
		},
		{
			name:   "PAID - sets NumberOfSuccessfulCharges to 2",
			status: constant.UnifiedPaymentSessionStatusPaid,
			payload: &AutoSplitPaymentSummary{
				NumberOfSuccessfulCharges: 1,
			},
			expected: AutoSplitPaymentSummary{
				NumberOfSuccessfulCharges: 2,
			},
		},
		{
			name:   "CANCELLED - sets NumberOfFailedCharges to 1",
			status: constant.UnifiedPaymentSessionStatusCancelled,
			expected: AutoSplitPaymentSummary{
				NumberOfFailedCharges: 1,
			},
		},
		{
			name:   "EXPIRED - sets NumberOfExpiredCharges to 1",
			status: constant.UnifiedPaymentSessionStatusExpired,
			expected: AutoSplitPaymentSummary{
				NumberOfExpiredCharges: 1,
			},
		},
		{
			name:   "EXPIRED - sets NumberOfExpiredCharges to 3",
			status: constant.UnifiedPaymentSessionStatusExpired,
			payload: &AutoSplitPaymentSummary{
				NumberOfExpiredCharges: 2,
			},
			expected: AutoSplitPaymentSummary{
				NumberOfExpiredCharges: 3,
			},
		},
		{
			name:   "PROCESSING - sets NumberOfInProcessCharges to 1",
			status: constant.UnifiedPaymentSessionStatusProcessing,
			expected: AutoSplitPaymentSummary{
				NumberOfInProcessCharges: 1,
			},
		},
		{
			name:   "REQUIRE_ACTION - sets NumberOfInProcessCharges to 1",
			status: constant.UnifiedPaymentSessionStatusRequireAction,
			expected: AutoSplitPaymentSummary{
				NumberOfInProcessCharges: 1,
			},
		},
		{
			name:     "unknown status - no fields modified",
			status:   "UNKNOWN",
			expected: AutoSplitPaymentSummary{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AutoSplitPaymentSummary{}
			if tt.payload != nil {
				m = tt.payload
			}
			m.UpdateRecordByParentStatus(tt.status)
			assert.Equal(t, tt.expected, *m)
		})
	}
}
