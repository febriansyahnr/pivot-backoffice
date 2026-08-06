package internalXbController

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalXbController) GetListMasterCountry(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterCountry")
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

	requestPayload := xbModel.GetListMasterCountryRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	activeOnly := query.Get("activeOnly")
	if activeOnly != "" {
		requestPayload.ActiveOnly, err = strconv.ParseBool(activeOnly)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidActiveOnly))
			return
		}
	}

	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.CurrencyCode = query.Get("currencyCode")
	resp, err := c.xbPayoutSvc.GetListMasterCountry(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterCurrency(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterCurrency")
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

	requestPayload := xbModel.GetListMasterCurrencyRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterCurrency(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterState(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterState")
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

	requestPayload := xbModel.GetListMasterStateRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterState(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterCity(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterCity")
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

	requestPayload := xbModel.GetListMasterCityRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.StateUUID = query.Get("stateUuid")
	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterCity(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterCurrencyMapping(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterCurrencyMapping")
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

	requestPayload := xbModel.GetListMasterCurrencyMappingRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.TransferMethod = query.Get("transferMethod")
	resp, err := c.xbPayoutSvc.GetListMasterCurrencyMapping(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterIdentificationType(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterIdentificationType")
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

	requestPayload := xbModel.GetListMasterIdentificationTypeRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.AccountType = query.Get("accountType")
	resp, err := c.xbPayoutSvc.GetListMasterIdentificationType(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterAccountType(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterAccountType")
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

	requestPayload := xbModel.GetListMasterAccountTypeRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterAccountType(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterPurpose(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterPurpose")
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

	requestPayload := xbModel.GetListMasterPurposeRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.RoutingCode = query.Get("routingCode")
	resp, err := c.xbPayoutSvc.GetListMasterPurpose(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterTransferMethod(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterTransferMethod")
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

	requestPayload := xbModel.GetListMasterTransferMethodRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterTransferMethod(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetListMasterSourceOfIncome(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListMasterSourceOfIncome")
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

	requestPayload := xbModel.GetListMasterSourceOfIncomeRequest{
		MerchantId: merchantID,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterSourceOfIncome(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}
