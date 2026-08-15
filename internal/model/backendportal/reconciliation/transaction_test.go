package reconciliation_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestReconDetail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		detail  reconciliation.ReconDetail
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty status should return error",
			detail:  reconciliation.ReconDetail{Status: ""},
			wantErr: true,
			errMsg:  "status is required",
		},
		{
			name:    "invalid status should return error",
			detail:  reconciliation.ReconDetail{Status: "INVALID"},
			wantErr: true,
			errMsg:  constant.ErrInvalidStatus.Error(),
		},
		{
			name:    "lowercase failed status should be normalized",
			detail:  reconciliation.ReconDetail{Status: "review"},
			wantErr: false,
		},
		{
			name:    "uppercase failed status should be valid",
			detail:  reconciliation.ReconDetail{Status: "REVIEW"},
			wantErr: false,
		},
		{
			name:    "lowercase success status should be normalized",
			detail:  reconciliation.ReconDetail{Status: "true"},
			wantErr: false,
		},
		{
			name:    "uppercase success status should be valid",
			detail:  reconciliation.ReconDetail{Status: "TRUE"},
			wantErr: false,
		},
		{
			name:    "mixed case failed status should be normalized",
			detail:  reconciliation.ReconDetail{Status: "ReView"},
			wantErr: false,
		},
		{
			name:    "mixed case success status should be normalized",
			detail:  reconciliation.ReconDetail{Status: "TrUe"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.detail.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errMsg, err.Error())
			} else {
				assert.NoError(t, err)
				if tt.detail.Status == "review" || tt.detail.Status == "REVIEW" {
					assert.Equal(t, constant.ReconStatusReview, tt.detail.Status)
				} else {
					assert.Equal(t, constant.ReconStatusSuccess, tt.detail.Status)
				}
			}
		})
	}
}

func TestPaymentTotalAmountResult_Add(t *testing.T) {
	tests := []struct {
		name      string
		initial   reconciliation.PaymentTotalAmountResult
		reference string
		amount    decimal.Decimal
		expected  decimal.Decimal
	}{
		{
			name:      "add to empty map",
			initial:   make(reconciliation.PaymentTotalAmountResult),
			reference: "REF001",
			amount:    decimal.NewFromFloat(100.50),
			expected:  decimal.NewFromFloat(100.50),
		},
		{
			name: "add to existing reference",
			initial: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(50.25),
			},
			reference: "REF001",
			amount:    decimal.NewFromFloat(25.75),
			expected:  decimal.NewFromFloat(76.00),
		},
		{
			name: "add different reference to existing map",
			initial: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(100.00),
			},
			reference: "REF002",
			amount:    decimal.NewFromFloat(200.00),
			expected:  decimal.NewFromFloat(200.00),
		},
		{
			name:      "add zero amount",
			initial:   make(reconciliation.PaymentTotalAmountResult),
			reference: "REF001",
			amount:    decimal.Zero,
			expected:  decimal.Zero,
		},
		{
			name:      "add negative amount",
			initial:   make(reconciliation.PaymentTotalAmountResult),
			reference: "REF001",
			amount:    decimal.NewFromFloat(-50.25),
			expected:  decimal.NewFromFloat(-50.25),
		},
		{
			name: "add negative to positive amount",
			initial: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(100.00),
			},
			reference: "REF001",
			amount:    decimal.NewFromFloat(-30.00),
			expected:  decimal.NewFromFloat(70.00),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial
			result.Add(tt.reference, tt.amount)

			actual := result.GetTotalAmount(tt.reference)
			assert.True(t, tt.expected.Equal(actual), "Expected %s, got %s", tt.expected, actual)
		})
	}
}

func TestPaymentTotalAmountResult_GetTotalAmount(t *testing.T) {
	tests := []struct {
		name      string
		result    reconciliation.PaymentTotalAmountResult
		reference string
		expected  decimal.Decimal
	}{
		{
			name:      "get from empty map",
			result:    make(reconciliation.PaymentTotalAmountResult),
			reference: "REF001",
			expected:  decimal.Zero,
		},
		{
			name: "get existing reference",
			result: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(150.75),
				"REF002": decimal.NewFromFloat(200.00),
			},
			reference: "REF001",
			expected:  decimal.NewFromFloat(150.75),
		},
		{
			name: "get non-existing reference",
			result: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(100.00),
			},
			reference: "REF999",
			expected:  decimal.Zero,
		},
		{
			name: "get zero amount",
			result: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.Zero,
			},
			reference: "REF001",
			expected:  decimal.Zero,
		},
		{
			name: "get negative amount",
			result: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(-75.50),
			},
			reference: "REF001",
			expected:  decimal.NewFromFloat(-75.50),
		},
		{
			name: "get from map with multiple references",
			result: reconciliation.PaymentTotalAmountResult{
				"REF001": decimal.NewFromFloat(100.00),
				"REF002": decimal.NewFromFloat(250.75),
				"REF003": decimal.NewFromFloat(0.01),
			},
			reference: "REF002",
			expected:  decimal.NewFromFloat(250.75),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.result.GetTotalAmount(tt.reference)
			assert.True(t, tt.expected.Equal(actual), "Expected %s, got %s", tt.expected, actual)
		})
	}
}
