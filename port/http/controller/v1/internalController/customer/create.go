package customerController

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalCustomerController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/Create")
	defer segment.End()

	var (
		err error
	)

	var payload customerModel.CreateCustomerRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	merchantInfo := ctx.Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}
	payload.MerchantID = merchant.MerchantId

	err = c.validate.Struct(payload)
	if err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.customerService.CreateCustomer(ctx, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}

func (c *V1InternalCustomerController) CreateWalletCustomer(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/CreateWalletCustomer")
	defer segment.End()

	var (
		err error
	)

	var payload customerModel.CreateCustomerRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	httputil.BindMerchantID(r, &payload.MerchantID)
	err = c.validate.Struct(payload)
	if err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.customerService.CreateCustomer(ctx, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
