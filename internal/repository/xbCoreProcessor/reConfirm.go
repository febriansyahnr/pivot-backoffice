package xbCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *xbCoreProcessorRepository) ReConfirmPayout(ctx context.Context, request *xbCoreProcessorModel.ConfirmPayoutRequest) (xbCoreProcessorModel.ReConfirmPayoutResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/ConfirmPayout")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout/%s/reconfirm", r.config.XbCoreProcessorConfig.BaseUrl, request.XbPayoutId)
	r.logger.Info(ctx, "ReConfirmPayout", logger.String("url", url))

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
		return xbCoreProcessorModel.ReConfirmPayoutResponse{}, err
	}

	var resp xbCoreProcessorModel.ReConfirmPayoutResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb confirm payout session body", logger.Error(err))
		return xbCoreProcessorModel.ReConfirmPayoutResponse{}, err
	}

	r.logger.Info(ctx, "ReConfirmPayout", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(string(response)))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when xb confirm payout session, errorCode %s", resp.Code), logger.Error(err))
		return resp, errors.New(string(response))
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when xb confirm payout session, errorCode %s", resp.Code), logger.Error(err))
		return resp, err
	}

	return resp, nil
}
