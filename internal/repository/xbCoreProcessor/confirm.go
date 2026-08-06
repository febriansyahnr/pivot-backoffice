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

func (r *xbCoreProcessorRepository) ConfirmPayout(ctx context.Context, request *xbCoreProcessorModel.ConfirmPayoutRequest) (*xbCoreProcessorModel.ConfirmPayoutResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/ConfirmPayout")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout/%s/confirm", r.config.XbCoreProcessorConfig.BaseUrl, request.XbPayoutId)
	r.logger.Info(ctx, "ConfirmPayout", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXMerchantId:         request.MerchantId,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request xb confirm payout session", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.ConfirmPayoutResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb confirm payout session body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "ConfirmPayout", logger.ByteString("response", response))

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb confirm payout session, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
