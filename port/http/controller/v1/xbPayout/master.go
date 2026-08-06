package xbPayoutController

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *xbPayoutController) GetListMasterCountry(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterCountry")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterCountryRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	activeOnly := query.Get("activeOnly")
	if activeOnly != "" {
		requestPayload.ActiveOnly, err = strconv.ParseBool(activeOnly)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidActiveOnly))
			return
		}
	}

	resp, err := c.xbPayoutSvc.GetListMasterCountry(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterCurrency(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterCurrency")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterCurrencyRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterCurrency(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterState(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterState")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterStateRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterState(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterCity(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterCity")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterCityRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.StateUUID = query.Get("stateUuid")
	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterCity(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterCurrencyMapping(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterCurrencyMapping")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterCurrencyMappingRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.TransferMethod = query.Get("transferMethod")
	resp, err := c.xbPayoutSvc.GetListMasterCurrencyMapping(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterIdentificationType(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterIdentificationType")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterIdentificationTypeRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.AccountType = query.Get("accountType")
	resp, err := c.xbPayoutSvc.GetListMasterIdentificationType(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterAccountType(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterAccountType")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterAccountTypeRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterAccountType(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterPurpose(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterPurpose")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterPurposeRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.RoutingCode = query.Get("routingCode")
	resp, err := c.xbPayoutSvc.GetListMasterPurpose(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterTransferMethod(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterTransferMethod")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterTransferMethodRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Code = query.Get("code")
	resp, err := c.xbPayoutSvc.GetListMasterTransferMethod(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *xbPayoutController) GetListMasterSourceOfIncome(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetListMasterSourceOfIncome")
	defer segment.End()

	var err error

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	requestPayload := xbModel.GetListMasterSourceOfIncomeRequest{
		MerchantId: user.MerchantId,
		Page:       constant.DefaultPage,
		PerPage:    constant.DefaultPageSize,
	}

	query := r.URL.Query()

	page := query.Get("page")
	if page != "" {
		requestPayload.Page, err = strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}

	perPage := query.Get("perPage")
	if perPage != "" {
		requestPayload.PerPage, err = strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}

	fetchAll := query.Get("fetchAll")
	if fetchAll != "" {
		requestPayload.FetchAll, err = strconv.ParseBool(fetchAll)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFetchAll))
			return
		}
	}

	requestPayload.Name = query.Get("name")
	resp, err := c.xbPayoutSvc.GetListMasterSourceOfIncome(ctx, &requestPayload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}
