package internalXbController

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalXbController) SubmitRfiDetails(w http.ResponseWriter, r *http.Request) {
	var (
		err     error
		request *xbModel.SubmitRfiDetailsRequest
		ctx     = r.Context()
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/xb/SubmitRfiDetails")
	defer segment.End()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	payoutId := chi.URLParam(r, "id")
	if err := uuid.Validate(payoutId); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	documentId := r.FormValue("documentId")
	if err := uuid.Validate(documentId); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrDocumentIdIsRequired))
		return
	}

	request = &xbModel.SubmitRfiDetailsRequest{
		PayoutId:   payoutId,
		MerchantId: merchantId,
		DocumentId: documentId,
		Comment:    r.FormValue("comment"),
		Value:      r.FormValue("value"),
	}

	if err := c.validate.Struct(request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, err))
		return
	}

	if _, request.Document, err = r.FormFile("document"); err != nil && err != http.ErrMissingFile {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	// Value and document cannot be exist at the same time
	if request.Value != "" && request.Document != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("value and document cannot be exist at the same time")))
		return
	}

	// Either value or document must be exist
	if request.Value == "" && request.Document == nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("either value or document must be exist")))
		return
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameXB,
			MetricName:           constant.MetricNameXBSubmitRFI,
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			Attributes: map[string]any{
				"merchantId":          merchantCtx.MerchantId,
				"onBehalfSubmerchant": merchantCtx.MerchantId != merchantId,
			},
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "failed to record xb upload payout document custom metric", logger.Error(errMetric))
		}
	}()

	rfiDetails, err := c.xbPayoutSvc.SubmitRfiDetails(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, rfiDetails)
}
