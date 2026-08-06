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

func (r *xbCoreProcessorRepository) CreateBeneficiary(
	ctx context.Context,
	request *xbCoreProcessorModel.CreateBeneficiaryRequest,
) (*xbCoreProcessorModel.CreateBeneficiaryData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/CreateBeneficiary")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-beneficiary", r.config.XbCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "CreateBeneficiary", logger.String("url", url), logger.Any("request", request))

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
		err = pkgErrors.New(httpResponse.HttpErrRequest, err)
		r.logger.Error(ctx, "error when do request xb create beneficiary", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "CreateBeneficiary", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateBeneficiaryResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, err)
		r.logger.Error(ctx, "error when read xb create beneficiary", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb create beneficiary, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListBeneficiary(ctx context.Context, request *xbCoreProcessorModel.GetListBeneficiaryRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListBeneficiary")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-beneficiary/list?&page=%d&per_page=%d&fetch_all=%t&show_deactivated=%t&name=%s&country_code=%s&account_number=%s&account_type=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.Page,
		request.PerPage,
		request.FetchAll,
		request.ShowDeactivated,
		request.Name,
		request.CountryCode,
		request.AccountNumber,
		request.AccountType,
	)
	r.logger.Info(ctx, "GetListBeneficiary", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb get list beneficiary", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListBeneficiary", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb get list beneficiary", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb get list beneficiary, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetBeneficiaryById(ctx context.Context, request *xbCoreProcessorModel.GetBeneficiaryByIdRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetBeneficiaryById")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-beneficiary/%s", r.config.XbCoreProcessorConfig.BaseUrl, request.BeneficiaryId)
	r.logger.Info(ctx, "GetBeneficiaryById", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb get beneficiary by id", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetBeneficiaryById", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateBeneficiaryResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb get beneficiary by id", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb get beneficiary by id, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) UpdateBeneficiary(ctx context.Context, request *xbCoreProcessorModel.UpdateBeneficiaryRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/UpdateBeneficiary")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-beneficiary/%s/update", r.config.XbCoreProcessorConfig.BaseUrl, request.BeneficiaryId)
	r.logger.Info(ctx, "UpdateBeneficiary", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb update beneficiary", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UpdateBeneficiary", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateBeneficiaryResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb update beneficiary", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb update beneficiary, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) DeactivateBeneficiary(ctx context.Context, request *xbCoreProcessorModel.GetBeneficiaryByIdRequest) (*xbCoreProcessorModel.CreateBeneficiaryData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/DeactivateBeneficiary")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout-beneficiary/%s/deactivate", r.config.XbCoreProcessorConfig.BaseUrl, request.BeneficiaryId)
	r.logger.Info(ctx, "DeactivateBeneficiary", logger.String("url", url), logger.Any("request", request))

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
		r.logger.Error(ctx, "error when do request xb deactivate beneficiary", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "DeactivateBeneficiary", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.CreateBeneficiaryResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb deactivate beneficiary", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb deactivate beneficiary, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
