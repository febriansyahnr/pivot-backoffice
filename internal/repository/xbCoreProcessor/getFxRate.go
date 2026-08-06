package xbCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *xbCoreProcessorRepository) GetFxRate(ctx context.Context, request *xbCoreProcessorModel.GetFxRateRequest) (*xbCoreProcessorModel.GetFxRateResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetFxRate")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/fx/rate?source_currency=%s&destination_currency=%s&type=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.SourceCurrency,
		request.DestinationCurrency,
		request.RequestType,
	)
	r.logger.Info(ctx, "GetFxRate", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXMerchantId:         request.MerchantId,
			constant.HeaderXRequestId:          requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get fx rate", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.GetFxRateResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get fx rate response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetFxRate", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get fx rate, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
