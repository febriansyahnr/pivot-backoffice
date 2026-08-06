package internalPaymentController

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// isSnapMultiplePaymentDelegationEnabled checks if the feature flag for delegating
// MULTIPLE payment types from Snap API to Unified Payment V2 is enabled
func (h *InternalPaymentController) isSnapMultiplePaymentDelegationEnabled(merchantAuth *merchant.MerchantAuthTokenClaims) bool {
	// Create evaluation context with merchant ID and environment
	ffContext := ffcontext.NewEvaluationContext(merchantAuth.MerchantId)
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameMerchantId, merchantAuth.MerchantId)

	// Use BoolVariation to check the feature flag (defaults to false for safety)
	enabled, _ := ffclient.BoolVariation(constant.FeatureFlagSnapMultiplePaymentDelegation, ffContext, false)
	return enabled
}

// getUnifiedPaymentSession retrieves a Unified Payment V2 session by payment session ID
func (h *InternalPaymentController) getUnifiedPaymentSession(ctx context.Context, merchantId, paymentSessionId string) (*unifiedPaymentModel.UnifiedPaymentSessionResponse, error) {
	if h.unifiedPaymentSvc == nil {
		return nil, nil
	}

	request := &unifiedPaymentModel.GetUnifiedPaymentSessionRequest{
		PaymentSessionID: paymentSessionId,
		MerchantID:       merchantId,
	}
	
	return h.unifiedPaymentSvc.GetSessionDetail(ctx, request)
}

// mapUnifiedStatusToSnapStatus maps Unified Payment V2 statuses to Snap API statuses
func (h *InternalPaymentController) mapUnifiedStatusToSnapStatus(unifiedStatus string) string {
	switch unifiedStatus {
	case "ACTIVE":
		return "PENDING"
	case "INACTIVE", "CANCELLED":
		return "CANCELLED"
	case "COMPLETED":
		return "SUCCESS"
	case "EXPIRED":
		return "EXPIRED"
	default:
		return "PENDING"
	}
}
