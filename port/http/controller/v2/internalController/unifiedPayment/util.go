package v2InternalUnifiedPaymentController

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func (c *paymentController) validateAmountRange(amount float64, minAmount, maxAmount *float64) error {
	if minAmount != nil && *minAmount > amount {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentBelowMinAmount)
	}
	if maxAmount != nil && *maxAmount < amount {
		return pkgErrors.New(response.HttpErrRequest, constant.ErrPaymentAboveMaxAmount)
	}
	return nil
}

func (c *paymentController) isCybersourceTestAmountAllowed(merchantID string, amount float64) bool {
	if merchantID == "" || c.config.Environment != constant.EnvironmentStaging {
		return false
	}

	ffContext := ffcontext.NewEvaluationContext(merchantID)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantID)
	ffContext.AddCustomAttribute("amount", amount)

	result, _ := ffclient.BoolVariation(constant.FeatureFlagCybersourceTestAmountWhitelist, ffContext, false)
	return result
}
