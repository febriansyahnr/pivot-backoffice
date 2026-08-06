package internalXbController

import (
	"encoding/json"
	"net/http"

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

func (c *InternalXbController) CreatePayoutSession(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/CreatePayoutSession")
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

	// Find merchant
	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantID)
	if err != nil || merchant == nil {
		errMerchantNotFound := pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantNotFound)
		response.SendOpenApiNonSnapResponseError(ctx, w, errMerchantNotFound)
		return
	}

	requestPayload := xbModel.CreatePayoutSessionRequest{
		MerchantId:   merchantID,
		MerchantName: merchant.Name,
		CreatedBy:    merchantCtx.MerchantId,
		CreatedFrom:  constant.DisbursementCreatedFromOpenApi,
	}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		err = pkgErrors.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload)

		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if err := c.validateCreateRequest(requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	defer func() {
		metricData := &monitoring.CustomMetric{
			ComponentName:        constant.ComponentNameXB,
			MetricName:           constant.MetricNameXBCreatePayout,
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
			c.logger.Error(ctx, "failed to record xb create payout session custom metric", logger.Error(errMetric))
		}
	}()

	resp, err := c.xbPayoutSvc.CreateSession(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) validateCreateRequest(req xbModel.CreatePayoutSessionRequest) error {
	// First, validate the main struct fields
	err := c.validate.StructExcept(req, "SenderData", "BeneficiaryData")
	if err != nil {
		return err
	}

	// Conditionally validate SenderData if SenderID is empty
	if req.SenderID == "" && req.SenderData != nil {
		err = c.validate.Struct(req.SenderData)
		if err != nil {
			return err
		}
	}

	// Conditionally validate BeneficiaryData if BeneficiaryID is empty
	if req.BeneficiaryID == "" && req.BeneficiaryData != nil {
		err = c.validate.Struct(req.BeneficiaryData)
		if err != nil {
			return err
		}
	}

	return nil
}
