package customerController

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *V1InternalCustomerController) Delete(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/customer/Delete")
	defer segment.End()

	var (
		err error
	)

	customerId := chi.URLParam(r, "id")
	err = c.validate.Var(customerId, "required,uuid")
	if err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	merchantInfo := ctx.Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	_, err = c.customerService.DeleteCustomer(ctx, customerId, merchant.MerchantId)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendGeneralResponseOK(w, map[string]interface{}{"deleted": true})
}
