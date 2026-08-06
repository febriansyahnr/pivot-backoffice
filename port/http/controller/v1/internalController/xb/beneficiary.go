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

func (c *InternalXbController) CreateBeneficiary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/CreateBeneficiary")
	defer segment.End()

	// Merchant info from JWT
	merchantCtx, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantId := merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantId)

	requestPayload := xbModel.CreateBeneficiaryRequest{
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

	resp, err := c.xbPayoutSvc.CreateBeneficiary(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) GetListBeneficiary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetListBeneficiary")
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

	requestPayload := xbModel.GetListBeneficiaryRequest{
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
	requestPayload.CountryCode = query.Get("countryCode")
	requestPayload.AccountNumber = query.Get("accountNumber")
	requestPayload.AccountType = query.Get("accountType")

	resp, err := c.xbPayoutSvc.GetListBeneficiary(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, resp.Results, resp.Pagination)
}

func (c *InternalXbController) GetBeneficiaryById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/GetBeneficiaryById")
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

	resp, err := c.xbPayoutSvc.GetBeneficiaryById(ctx, &xbModel.GetBeneficiaryByIdRequest{
		MerchantId:    merchantID,
		BeneficiaryId: id,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) UpdateBeneficiary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/UpdateBeneficiary")
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

	requestPayload := xbModel.UpdateBeneficiaryRequest{
		BeneficiaryId: id,
		MerchantId:    merchantId,
	}

	if err := json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(requestPayload); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.xbPayoutSvc.UpdateBeneficiary(ctx, &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}

func (c *InternalXbController) DeactivateBeneficiary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/xb/DeactivateBeneficiary")
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

	resp, err := c.xbPayoutSvc.DeactivateBeneficiary(ctx, &xbModel.GetBeneficiaryByIdRequest{
		MerchantId:    merchantId,
		BeneficiaryId: id,
	})
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, resp)
}
