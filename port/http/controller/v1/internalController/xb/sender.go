package internalXbController

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *InternalXbController) CreateSender(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/CreateSender")
	defer segment.End()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	requestPayload := xbModel.CreateSenderRequest{
		MerchantId: merchantId,
	}
	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.xbPayoutSvc.CreateSender(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) GetListSender(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListSender")
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

	requestPayload := xbModel.GetListSenderRequest{
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

	showDeactivated := query.Get("showDeactivated")
	if showDeactivated != "" {
		requestPayload.ShowDeactivated, err = strconv.ParseBool(showDeactivated)
		if err != nil {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidShowDeactivated))
			return
		}
	}

	requestPayload.Name = query.Get("name")
	requestPayload.AccountType = query.Get("accountType")

	resp, err := c.xbPayoutSvc.GetListSender(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetSenderById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetSenderById")
	defer segment.End()

	var err error

	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	resp, err := c.xbPayoutSvc.GetSenderById(ctx, &xbModel.GetSenderByIdRequest{
		MerchantId: merchantID,
		SenderId:   id,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) UpdateSender(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/UpdateSender")
	defer segment.End()

	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	requestPayload := xbModel.UpdateSenderRequest{
		SenderId:   id,
		MerchantId: merchantId,
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.xbPayoutSvc.UpdateSender(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) DeactivateSender(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/DeactivateSender")
	defer segment.End()

	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	resp, err := c.xbPayoutSvc.DeactivateSender(ctx, &xbModel.GetSenderByIdRequest{
		MerchantId: merchantId,
		SenderId:   id,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}
