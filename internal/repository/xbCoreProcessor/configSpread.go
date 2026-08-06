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

func (r *xbCoreProcessorRepository) GetListConfigSpread(ctx context.Context, request *xbCoreProcessorModel.GetListConfigSpreadRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListConfigSpread")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/config/spread/list?merchant_id=%s&page=%d&per_page=%d",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.MerchantID,
		request.Page,
		request.PerPage,
	)
	r.logger.Info(ctx, "GetListConfigSpread", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get config spread list", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get config spread list response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListConfigSpread", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get config spread list, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetConfigSpreadDetailByID(ctx context.Context, id string) (*xbCoreProcessorModel.GetConfigSpreadDetailData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetConfigSpreadDetailByID")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/config/spread/%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		id,
	)
	r.logger.Info(ctx, "GetConfigSpreadDetailByID", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get config spread detail", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.GetConfigSpreadDetailResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get config spread detail response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetConfigSpreadDetailByID", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get config spread detail, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) CreateConfigSpread(ctx context.Context, request *xbCoreProcessorModel.CreateConfigSpreadRequest) (*xbCoreProcessorModel.CreateConfigSpreadData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/CreateConfigSpread")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/config/spread", r.config.XbCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "CreateConfigSpread", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request xb create config spread", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.CreateConfigSpreadResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb create config spread body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CreateConfigSpread", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb create config spread, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) UpdateConfigSpread(ctx context.Context, request *xbCoreProcessorModel.UpdateConfigSpreadRequest) (*xbCoreProcessorModel.UpdateConfigSpreadData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/UpdateConfigSpread")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/config/spread/%s", r.config.XbCoreProcessorConfig.BaseUrl, request.UUID.String())
	r.logger.Info(ctx, "UpdateConfigSpread", logger.String("url", url))

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.PUT(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request xb update config spread", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.UpdateConfigSpreadResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb update config spread body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UpdateConfigSpread", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb update config spread, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
