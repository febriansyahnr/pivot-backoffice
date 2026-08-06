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

func (r *xbCoreProcessorRepository) GetPayoutById(ctx context.Context, request *xbCoreProcessorModel.GetPayoutRequest) (*xbCoreProcessorModel.GetPayoutResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetPayoutById")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout/%s", r.config.XbCoreProcessorConfig.BaseUrl, request.Id)
	r.logger.Info(ctx, "GetPayoutById", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXMerchantId:         request.MerchantId,
			constant.XRequestIdKey:             requestId,
		},
	)

	if err != nil {
		r.logger.Error(ctx, "error when do request get payout by id", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.GetPayoutResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get payout by id response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetPayoutById", logger.ByteString("response", response))

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, resp.Message)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get payout by id, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
