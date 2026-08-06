package simulationController

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
)

// GetRedirectionURL retrieves the appropriate redirection URL for a unified payment based on charge status.
//
// This function fetches payment details by ID and returns the corresponding redirect URL
// based on the request charge status (paid, cancelled, or expired). It only processes
// payments that are in API mode (UnifiedPaymentModeAPI).
//
// Behavior:
//   - Returns empty string (no error) if payment is not in API mode
//   - Returns constant.ErrPaymentNotFound if payment doesn't exist
//   - Maps payment status to corresponding redirect URL:
//   - UnifiedPaymentSessionStatusPaid -> SuccessReturnUrl
//   - UnifiedPaymentSessionStatusCancelled -> FailureReturnUrl
//   - UnifiedPaymentSessionStatusExpired -> ExpirationReturnUrl
func (h *Handler) GetRedirectionURL(ctx context.Context, paymentID string, ChargeStatus string) (string, error) {
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/simulation/GetRedirectionURL")
	defer segment.End()
	var url = ""
	payment, err := h.paymentSvc.GetDetailByID(ctx, paymentID)
	if err != nil {
		return url, err
	}

	if payment == nil {
		return url, constant.ErrPaymentNotFound
	}

	// mode redirect always use same url and the ui will mapping the next item
	if payment.Metadata == nil || (*payment.Metadata)["mode"] != constant.UnifiedPaymentModeAPI {
		return url, nil
	}

	redirectURL, err := util.ConvertToStruct[unifiedPaymentModel.RedirectUrl]((*payment.Metadata)["clientRedirectUrl"])
	if err != nil {
		return url, err
	}

	// cannot use payment status due to latency in other services
	switch ChargeStatus {
	case constant.ChargeStatusHistorySuccess:
		url = redirectURL.SuccessReturnUrl
	case constant.ChargeStatusHistoryFailed:
		url = redirectURL.FailureReturnUrl
	case constant.ChargeStatusHistoryExpired:
		url = redirectURL.ExpirationReturnUrl
	}

	return url, nil
}
