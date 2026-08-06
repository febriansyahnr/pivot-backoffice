package xbCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *xbCoreProcessorRepository) CreateSender(
	ctx context.Context,
	request *xbCoreProcessorModel.CreateSenderRequest,
) (*xbCoreProcessorModel.CreateSenderData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/CreateSender")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-remitter", r.config.XbCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "CreateSender", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb create sender", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CreateSender", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateSenderResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, err)
		r.logger.Error(ctx, "error when read xb create sender", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb create sender, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListSender(ctx context.Context, request *xbCoreProcessorModel.GetListSenderRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListSender")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-remitter/list?&page=%d&per_page=%d&fetch_all=%t&show_deactivated=%t&name=%s&account_type=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.Page,
		request.PerPage,
		request.FetchAll,
		request.ShowDeactivated,
		request.Name,
		request.AccountType,
	)
	r.logger.Info(ctx, "GetListSender", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb get list sender", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListSender", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb get list sender", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb get list sender, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetSenderById(ctx context.Context, request *xbCoreProcessorModel.GetSenderByIdRequest) (*xbCoreProcessorModel.CreateSenderData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetSenderById")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-remitter/%s", r.config.XbCoreProcessorConfig.BaseUrl, request.SenderId)
	r.logger.Info(ctx, "GetSenderById", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb get sender by id", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetSenderById", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateSenderResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb get sender by id", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb get sender by id, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) UpdateSender(ctx context.Context, request *xbCoreProcessorModel.UpdateSenderRequest) (*xbCoreProcessorModel.CreateSenderData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/UpdateSender")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-remitter/%s/update", r.config.XbCoreProcessorConfig.BaseUrl, request.SenderId)
	r.logger.Info(ctx, "UpdateSender", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb update sender", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UpdateSender", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateSenderResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb update sender", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb update sender, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) DeactivateSender(ctx context.Context, request *xbCoreProcessorModel.GetSenderByIdRequest) (*xbCoreProcessorModel.CreateSenderData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/DeactivateSender")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-remitter/%s/deactivate", r.config.XbCoreProcessorConfig.BaseUrl, request.SenderId)
	r.logger.Info(ctx, "DeactivateSender", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb deactivate sender", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "DeactivateSender", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateSenderResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb deactivate sender", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb deactivate sender, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
