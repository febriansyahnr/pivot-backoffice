package customerController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalCustomerController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/GetList")
	defer segment.End()

	var (
		err error
	)

	// get merchant id from query params
	merchantId := r.URL.Query().Get("merchant_id")
	err = c.validate.Var(merchantId, "required,uuid")
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	page, perPage := httputil.GetPaginationQuery(r)

	phoneNumber := r.URL.Query().Get("phoneNumber")
	if phoneNumber != "" {
		err := c.validate.Var(phoneNumber, "numeric")
		if err != nil {
			response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
	}

	resp, err := c.customerService.GetCustomerList(ctx, merchantId, phoneNumber, page, perPage)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (c *V1InternalCustomerController) GetById(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/GetById")
	defer segment.End()

	var (
		err error
	)

	customerId := chi.URLParam(r, "id")
	err = c.validate.Var(customerId, "required,uuid")
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.customerService.FindCustomerByID(ctx, customerId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}

func (c *V1InternalCustomerController) GetByPhoneNumber(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/GetByPhoneNumber")
	defer segment.End()

	var (
		err error
	)

	phoneNumber := chi.URLParam(r, "phoneNumber")
	err = c.validate.Var(phoneNumber, "required")
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	merchantId, _ := ctx.Value(constant.CtxMerchantIDKey).(string)
	err = c.validate.Var(merchantId, "required,uuid")
	if err != nil {
		response.SendGeneralResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.customerService.GetCustomerByPhoneNumber(ctx, phoneNumber, merchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}

// TODO: We need to tidy up the routes for routes that use Authentication Token Middleware
// and routes that use Internal Service Middleware
func (c *V1InternalCustomerController) GetByIDForUnifiedPayment(w http.ResponseWriter, r *http.Request) {
	ctx, span := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/GetByIDForUnifiedPayment")
	defer span.End()

	var (
		err error
	)

	customerId := chi.URLParam(r, "id")
	err = c.validate.Var(customerId, "required,uuid")
	if err != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}

	merchantInfo := ctx.Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, err))
		return
	}

	resp, err := c.customerService.GetCustomerByIDForUnifiedPayment(ctx, customerId, merchant.MerchantId)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
