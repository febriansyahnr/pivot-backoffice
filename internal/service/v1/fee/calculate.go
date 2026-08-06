package feeService

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	"github.com/shopspring/decimal"
)

func (s *FeeService) CalculateFee(ctx context.Context, request *feeModel.GetFeeRequest, feeDetail *feeModel.FeeMetadataObject) (fee, tax float64) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/fee/CalculateFee")
	defer segment.End()

	// Note: In this section, only the ReferenceAmount attribute in the request arguments is used, other values ​​are ignored.

	// define amount
	switch feeDetail.AmountType {
	case constant.MerchantFeePercentageType:
		fee = (feeDetail.Percentage / 100) * request.ReferenceAmount // usecase amount * percentage
		if feeDetail.MaxFeeAmount != nil && fee > *feeDetail.MaxFeeAmount {
			fee = *feeDetail.MaxFeeAmount
		}

	case constant.MerchantFeeAmountPercentageType:
		fee = feeDetail.Amount + ((feeDetail.Percentage / 100) * request.ReferenceAmount) // fee amount + (usecase amount * percentage)

	default:
		fee = feeDetail.Amount
	}

	// calculate tax
	if feeDetail.TaxType == constant.MerchantTaxTypeExclusive {
		tax = (feeDetail.TaxPercentage / 100) * fee // fee amount * tax percentage
		fee += tax                                  // fee amount + tax

	} else if feeDetail.TaxType == constant.MerchantTaxTypeInclusive {
		dpp := (100 / (100 + feeDetail.TaxPercentage)) * fee // dpp = (100 / (100 + taxPercentage)) * fee
		tax = (feeDetail.TaxPercentage / 100) * dpp          // ppn = (taxPercentage / 100) * dpp
	}

	if fee > 0.00 {
		fee = decimal.NewFromFloat(fee).Round(0).InexactFloat64()
	}
	return fee, tax
}
