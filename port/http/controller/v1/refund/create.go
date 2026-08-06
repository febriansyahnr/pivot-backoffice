package refund

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// Create godoc
// @Summary		Create Refund Endpoint
// @Description	This endpoint is used to initiate a refund from the merchant dashboard
// @ID			api-refund-create-dashboard
// @Tags		API - Refund
// @Accept		json
// @Produce		json
// @Success		200  	{object}	response.ApiResponse
// @Failure		400  	{object}	response.ApiResponse
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/refunds [post]
// @Security	Bearer
func (c *RefundController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/refund/Create")
	defer span.End()

	var (
		resp *refundModel.RefundResponse
		err  error
	)

	// Get user info from JWT token (dashboard auth)
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	merchantID := user.MerchantId

	payload := refundModel.NewCreatRefundRequest()
	payload.MerchantID = merchantID
	payload.IsCRMRequest = false

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentRefund,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId": merchantID,
				"method":     payload.Method,
				"source":     "dashboard",
			},
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "failed to record refund create custom metric", logger.Error(errMetric))
		}
	}()

	if resp, err = c.refundService.Create(ctx, payload); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
