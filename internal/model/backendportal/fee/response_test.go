package feeModel

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/go/util"

	"github.com/stretchr/testify/assert"
)

func TestFeeMetadataObjectWithTotalAmountResponse(t *testing.T) {
	// Test cases
	tests := []struct {
		name        string
		feeDetail   FeeMetadataObject
		totalAmount float64
		expected    FeeResponse
	}{
		{
			name: "Basic fee with total amount",
			feeDetail: FeeMetadataObject{
				Type:          "TRANSACTION",
				Method:        "PERCENTAGE",
				DeductionType: "AUTOMATED",
				AmountType:    "FIXED",
				Amount:        5000,
				Percentage:    0.5,
				TaxType:       "VAT",
				TaxPercentage: 11,
				TaxAmount:     550,
			},
			totalAmount: 10000,
			expected: FeeResponse{
				Type:          "TRANSACTION",
				Method:        "PERCENTAGE",
				DeductionType: "AUTOMATED",
				AmountType:    "FIXED",
				Amount:        5000,
				Percentage:    0.5,
				TaxType:       "VAT",
				TaxPercentage: 11,
				TaxAmount:     550,
				TotalAmount:   10000,
			},
		},
		{
			name: "Fee with max fee amount and deduction day",
			feeDetail: FeeMetadataObject{
				Type:          "TRANSACTION",
				Method:        "PERCENTAGE",
				DeductionType: "DIRECT",
				DeductionDay:  util.ValueToPtr[int16](15),
				AmountType:    "PERCENTAGE",
				Amount:        2500,
				MaxFeeAmount:  util.ValueToPtr[float64](10000),
				Percentage:    0.25,
				TaxType:       "VAT",
				TaxPercentage: 11,
				TaxAmount:     275,
			},
			totalAmount: 5000,
			expected: FeeResponse{
				Type:          "TRANSACTION",
				Method:        "PERCENTAGE",
				DeductionType: "DIRECT",
				DeductionDay:  util.ValueToPtr[int16](15),
				AmountType:    "PERCENTAGE",
				Amount:        2500,
				MaxFeeAmount:  util.ValueToPtr[float64](10000),
				Percentage:    0.25,
				TaxType:       "VAT",
				TaxPercentage: 11,
				TaxAmount:     275,
				TotalAmount:   5000,
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.feeDetail.WithTotalAmountResponse(tt.totalAmount)

			// Verify all fields are correctly mapped
			assert.Equal(t, tt.expected.Type, result.Type)
			assert.Equal(t, tt.expected.Method, result.Method)
			assert.Equal(t, tt.expected.DeductionType, result.DeductionType)
			assert.Equal(t, tt.expected.DeductionDay, result.DeductionDay)
			assert.Equal(t, tt.expected.AmountType, result.AmountType)
			assert.Equal(t, tt.expected.Amount, result.Amount)
			assert.Equal(t, tt.expected.MaxFeeAmount, result.MaxFeeAmount)
			assert.Equal(t, tt.expected.Percentage, result.Percentage)
			assert.Equal(t, tt.expected.TaxType, result.TaxType)
			assert.Equal(t, tt.expected.TaxPercentage, result.TaxPercentage)
			assert.Equal(t, tt.expected.TaxAmount, result.TaxAmount)
			assert.Equal(t, tt.expected.TotalAmount, result.TotalAmount)
		})
	}
}

func TestToFeeResponse(t *testing.T) {
	feeResponse := FeeResponse{
		Type:          constant.ReferenceDisbursement,
		DeductionType: constant.MerchantFeeDeductionTypeDirect,
		AmountType:    constant.MerchantFeeAmountType,
		Amount:        1_000,
		TaxType:       constant.MerchantTaxTypeNonPKP,
		TotalAmount:   1_000,
	}

	feeOnBehalf := &TrxFeeOnBehalfMetadata{
		Reference:   constant.ReferenceDisbursement,
		AmountType:  constant.MerchantFeeAmountType,
		Amount:      1_000,
		FinalAmount: 1_000,
	}
	assert.Equal(t, feeResponse, feeOnBehalf.ToFeeResponse())

	feeMerchant := &FeeMetadataObject{
		Type:          constant.ReferenceDisbursement,
		DeductionType: constant.MerchantFeeDeductionTypeDirect,
		AmountType:    constant.MerchantFeeAmountType,
		Amount:        1_000,
		TaxType:       constant.MerchantTaxTypeNonPKP,
		FinalAmount:   1_000,
	}
	assert.Equal(t, feeResponse, feeMerchant.ToFeeResponse())
}
