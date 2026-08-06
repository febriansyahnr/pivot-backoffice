package customer

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMCustomerController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/customer/Update")
	defer segment.End()

	customerID := chi.URLParam(r, "id")
	if customerID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrCustomerIDRequired))
		return
	}

	var request customerModel.UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidUnmarshalJSON))
		return
	}

	if request.MerchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDRequired))
		return
	}

	// Custom validation for isBlocked and blockReason
	if request.IsBlocked != nil && *request.IsBlocked && (request.BlockReason == nil || *request.BlockReason == "") {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrBlockReasonRequired))
		return
	}

	// Clear blockReason if isBlocked is false
	if request.IsBlocked != nil && !*request.IsBlocked {
		emptyReason := ""
		request.BlockReason = &emptyReason
	}

	request.UUID = customerID

	customer, err := c.customerService.UpdateCustomer(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, customer)
}