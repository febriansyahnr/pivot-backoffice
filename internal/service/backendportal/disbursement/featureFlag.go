package disbursementService

import (
	"context"
	"math"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

const defaultBulkValidateWorkers = 13

// feature flag for allow certain merchant to use beneficiary custom rule
// flag name: "backend-portal-merchant-allowed-beneficiary-payout-custom-rule"
func (s *DisbursementService) IsMerchantAllowedToUseBeneficiaryCustomRule(ctx context.Context, merchantId string, isCustomRule bool) bool {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/isMerchantAllowedToUseBeneficiaryCustomRule")
	defer segment.End()

	if !isCustomRule {
		return false
	}

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	result, err := ffclient.BoolVariation(constant.FeatureFlagKeyMerchantAllowedBeneficiaryPayoutCustomRule, attr, false)
	if err != nil {
		s.logger.Error(ctx, "failed to get backend-portal-merchant-allowed-beneficiary-payout-custom-rule flag", logger.Error(err))
	}
	return result
}

// feature flag for exclude certain merchant from beneficiary payout rules
// flag name: "backend-portal-merchant-allowed-exclude-beneficiary-payout-rules"
func (s *DisbursementService) IsMerchantAllowedExcludeBeneficiaryRules(ctx context.Context, merchantId string, amount float64) (maxAmount float64, isAllow bool) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/isMerchantAllowedExcludeBeneficiaryRules")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	result, err := ffclient.BoolVariation(constant.FeatureFlagKeyMerchantAllowedExcludeBeneficiaryPayoutRules, attr, false)
	if result {
		s.logger.Info(ctx, "merchant is allowed to exclude beneficiary payout rules", logger.String("merchantId", merchantId))
		// define max amount to bypass beneficiary payout rules
		// add 1 to avoid equal comparison
		if amount <= 0 {
			max := math.MaxInt
			maxAmount = float64(max)
		} else {
			maxAmount = amount + 1
		}
	}

	if err != nil {
		s.logger.Error(ctx, "failed to get backend-portal-merchant-allowed-exclude-beneficiary-payout-rules flag", logger.Error(err))
	}

	return maxAmount, result
}

// getBulkValidateWorkers returns the number of workers for bulk validate from Consul.
// Default is 13 workers if not configured or on error.
func (s *DisbursementService) getBulkValidateWorkers() int {
	attr := ffcontext.NewEvaluationContext(uuid.NewString())
	workers, err := ffclient.IntVariation(constant.FeatureFlagKeyBulkValidateWorkers, attr, defaultBulkValidateWorkers)
	if err != nil {
		s.logger.Warn(context.Background(), "failed to get bulk validate workers from consul, using default",
			logger.Error(err),
			logger.Int("default_workers", defaultBulkValidateWorkers))
		return defaultBulkValidateWorkers
	}

	// Validate workers count
	if workers <= 0 {
		s.logger.Warn(context.Background(), "invalid bulk validate workers from consul, using default",
			logger.Int("configured_workers", workers),
			logger.Int("default_workers", defaultBulkValidateWorkers))
		return defaultBulkValidateWorkers
	}

	return workers
}

func (s *DisbursementService) getOverbookingBankCodeListViaFlip(ctx context.Context, merchantId string) (result []string) {
	_, segment := otelTracer.Start(ctx, "internal/service/v1/disbursement/getOverbookingBankCodeListViaFlip")
	defer segment.End()

	attr := ffcontext.NewEvaluationContext(merchantId)
	attr.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantId)

	ffResults, _ := ffclient.JSONArrayVariation(constant.FeatureFlagKeyDisbursementOverbookingBankCodeListViaFlip, attr, []interface{}{})
	var raws []interface{}
	if len(ffResults) > 0 {
		firstResult := ffResults[0].(map[string]interface{})
		if bankCodes, ok := firstResult["bankCodes"].([]interface{}); ok {
			raws = bankCodes
		}
	}

	result = make([]string, len(raws))
	for i, raw := range raws {
		if _, ok := raw.(string); ok {
			result[i] = raw.(string)
		}
	}

	return
}
