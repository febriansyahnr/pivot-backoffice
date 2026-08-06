package internalXbController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *InternalXbController) ConfirmPayoutSession(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/ConfirmPayoutSession")
	defer segment.End()

	var err error
	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// Set context expose unmapping request error
	// ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	requestPayload := xbModel.ConfirmPayoutRequest{
		MerchantId: merchantID,
		PayoutId:   id,
		ApprovedBy: merchantCtx.MerchantId,
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameXB,
			MetricName:           constant.MetricNameXBConfirmPayout,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          merchantCtx.MerchantId,
				"onBehalfSubmerchant": merchantCtx.MerchantId != merchantID,
			},
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "failed to record xb confirm payout session custom metric", logger.Error(errMetric))
		}
	}()

	resp, err := c.xbPayoutSvc.Confirm(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
