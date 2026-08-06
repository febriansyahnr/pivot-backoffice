package customerController

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalCustomerController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/Update")
	defer segment.End()

	var (
		err error
	)

	var payload customerModel.UpdateCustomerRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	customerId := chi.URLParam(r, "id")
	err = c.validate.Var(customerId, "required,uuid")
	if err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}
	payload.UUID = customerId

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

	resp, err := c.customerService.UpdateCustomer(ctx, payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, resp)
}
