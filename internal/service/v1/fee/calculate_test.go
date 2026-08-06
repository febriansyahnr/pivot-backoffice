package feeService_test

import (
	"context"
	"fmt"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/fee"

	"github.com/stretchr/testify/assert"
)

func TestCalculateFee(t *testing.T) {
	feeSvc := New(nil, nil, nil)
	maxFeeAmount := float64(10)

	tests := []struct {
		request   *feeModel.GetFeeRequest
		feeDetail *feeModel.FeeMetadataObject
		wantFee   float64
		wantTax   float64
	}{
		{
			request: &feeModel.GetFeeRequest{
				ReferenceAmount: 1_000_000,
			},
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType: c.MerchantFeePercentageType,
				Percentage: 0.1,
			},
			wantFee: 1_000,
		},
		{
			request: &feeModel.GetFeeRequest{
				ReferenceAmount: 1_000_000,
			},
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:   c.MerchantFeePercentageType,
				Percentage:   0.1,
				MaxFeeAmount: &maxFeeAmount,
			},
			wantFee: 10,
		},
		{
			request: &feeModel.GetFeeRequest{
				ReferenceAmount: 1_380_000,
			},
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeePercentageType,
				Percentage:    0.1,
				TaxType:       c.MerchantTaxTypeExclusive,
				TaxPercentage: 11,
			},
			wantFee: 1_532, wantTax: 151.8,
		},
		{
			request: &feeModel.GetFeeRequest{
				ReferenceAmount: 1_000_000,
			},
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeeAmountPercentageType,
				Percentage:    0.1,
				Amount:        350,
				TaxType:       c.MerchantTaxTypeInclusive,
				TaxPercentage: 11,
			},
			wantFee: 1_350, wantTax: 133.783784,
		},
		{
			request: &feeModel.GetFeeRequest{
				ReferenceAmount: 4_380_000,
			},
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeeAmountPercentageType,
				Percentage:    0.1,
				Amount:        350,
				TaxType:       c.MerchantTaxTypeInclusive,
				TaxPercentage: 10,
			},
			wantFee: 4_730, wantTax: 430,
		},
		{
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType: c.MerchantFeeAmountType,
				Amount:     1_250,
			},
			wantFee: 1_250,
		},
		{
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeeAmountType,
				Amount:        450,
				TaxType:       c.MerchantTaxTypeExclusive,
				TaxPercentage: 11,
			},
			wantFee: 500, wantTax: 49.5,
		},
		{
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeeAmountType,
				Amount:        125,
				TaxType:       c.MerchantTaxTypeExclusive,
				TaxPercentage: 10,
			},
			wantFee: 138, wantTax: 12.5,
		},
		{
			feeDetail: &feeModel.FeeMetadataObject{
				AmountType:    c.MerchantFeeAmountType,
				Amount:        495.01,
				TaxType:       c.MerchantTaxTypeExclusive,
				TaxPercentage: 10,
			},
			wantFee: 545, wantTax: 49.501,
		},
	}
	for _, test := range tests {
		fee, tax := feeSvc.CalculateFee(context.Background(), test.request, test.feeDetail)

		assert.Equal(t, test.wantFee, fee)
		assert.Equal(t, fmt.Sprintf("%.6f", test.wantTax), fmt.Sprintf("%.6f", tax)) // Rounding to avoid unit test failures caused by processor architecture differences.
	}
}
