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

func (r *xbCoreProcessorRepository) GetListMasterCountry(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCountryRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterCountry")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/country/list?fetch_all=%t&page=%d&per_page=%d&active_only=%t&country_code=%s&currency_code=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.ActiveOnly,
		request.CountryCode,
		request.CurrencyCode,
	)
	r.logger.Info(ctx, "GetListMasterCountry", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data country", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data country response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterCountry", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data country, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterState(ctx context.Context, request *xbCoreProcessorModel.GetListMasterStateRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterState")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/state/list?fetch_all=%t&page=%d&per_page=%d&country_code=%s&name=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.CountryCode,
		request.Name,
	)
	r.logger.Info(ctx, "GetListMasterState", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data state", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data state response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterState", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data state, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterCity(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCityRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterCity")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/city/list?fetch_all=%t&page=%d&per_page=%d&state_uuid=%s&name=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.StateUUID,
		request.Name,
	)
	r.logger.Info(ctx, "GetListMasterCity", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data city", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data city response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterCity", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data city, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterCurrency(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCurrencyRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterCurrency")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/currency/list?fetch_all=%t&page=%d&per_page=%d&code=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.Code,
	)
	r.logger.Info(ctx, "GetListMasterCurrency", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data currency", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data currency response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterCurrency", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data currency, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterCurrencyMapping(ctx context.Context, request *xbCoreProcessorModel.GetListMasterCurrencyMappingRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterCurrencyMapping")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/currency/map/list?fetch_all=%t&page=%d&per_page=%d&country_code=%s&transfer_method=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.CountryCode,
		request.TransferMethod,
	)
	r.logger.Info(ctx, "GetListMasterCurrencyMapping", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data currency mapping", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data currency mapping response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterCurrencyMapping", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data currency mapping, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterIdentificationType(ctx context.Context, request *xbCoreProcessorModel.GetListMasterIdentificationTypeRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterIdentificationType")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/identification-type/list?fetch_all=%t&page=%d&per_page=%d&account_type=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.AccountType,
	)
	r.logger.Info(ctx, "GetListMasterIdentificationType", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data identification type", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data identification type response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterIdentificationType", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data identification type, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterAccountType(ctx context.Context, request *xbCoreProcessorModel.GetListMasterAccountTypeRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterAccountType")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/account-type/list?fetch_all=%t&page=%d&per_page=%d&code=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.Code,
	)
	r.logger.Info(ctx, "GetListMasterAccountType", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data account type", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data account type response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterAccountType", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data account type, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterPurpose(ctx context.Context, request *xbCoreProcessorModel.GetListMasterPurposeRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterPurpose")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/purpose/list?fetch_all=%t&page=%d&per_page=%d&code=%s&country_code=%s&routing_code=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.Code,
		request.CountryCode,
		request.RoutingCode,
	)
	r.logger.Info(ctx, "GetListMasterPurpose", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data purpose", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data purpose response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterPurpose", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data purpose, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterTransferMethod(ctx context.Context, request *xbCoreProcessorModel.GetListMasterTransferMethodRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterTransferMethod")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/transfer-method/list?fetch_all=%t&page=%d&per_page=%d&code=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.Code,
	)
	r.logger.Info(ctx, "GetListMasterTransferMethod", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data transfer method", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data transfer method response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterTransferMethod", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data transfer method, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}

func (r *xbCoreProcessorRepository) GetListMasterSourceOfIncome(ctx context.Context, request *xbCoreProcessorModel.GetListMasterSourceOfIncomeRequest) (*xbCoreProcessorModel.PaginationData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/GetListMasterSourceOfIncomeRequest")
	defer segment.End()

	url := fmt.Sprintf(
		"%s/api/v1/master/source-of-income/list?fetch_all=%t&page=%d&per_page=%d&name=%s",
		r.config.XbCoreProcessorConfig.BaseUrl,
		request.FetchAll,
		request.Page,
		request.PerPage,
		request.Name,
	)
	r.logger.Info(ctx, "GetListMasterSourceOfIncome", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request get master data source of income", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.PaginationResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get master data source of income response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetListMasterSourceOfIncome", logger.ByteString("response", response))


	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when get master data source of income, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
