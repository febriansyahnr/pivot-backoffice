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

func (r *xbCoreProcessorRepository) GetRfiDetails(ctx context.Context, request *xbCoreProcessorModel.GetRfiDetailsRequest) ([]*xbCoreProcessorModel.GetRfiDetailsResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetRfiDetails")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/rfi/%s", r.config.XbCoreProcessorConfig.BaseUrl, request.Id)
	r.logger.Info(ctx, "GetRfiDetails", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get rfi details", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetRfiDetails", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.GetRfiDetailsResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get rfi details response body", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get rfi details, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return resp.Data, nil
}
