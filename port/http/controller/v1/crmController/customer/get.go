package customer

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMCustomerController) GetCustomerByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/customer/GetCustomerByID")
	defer segment.End()

	merchantID := r.URL.Query().Get("merchant_id")
	customerID := chi.URLParam(r, "id")

	if merchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDRequired))
		return
	}

	if customerID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrCustomerIDRequired))
		return
	}

	customer, err := c.customerService.GetCustomerById(ctx, customerID, merchantID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, customer)
}