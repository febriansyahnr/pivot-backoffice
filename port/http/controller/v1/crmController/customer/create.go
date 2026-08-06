package customer

import (
	"encoding/json"
	"net/http"

	customerModel "github.com/paper-indonesia/pivot-backoffice/internal/model/customer"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMCustomerController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/customer/Create")
	defer segment.End()

	var request customerModel.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrInvalidUnmarshalJSON))
		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if request.MerchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDRequired))
		return
	}

	// Custom validation for isBlocked and blockReason
	if request.IsBlocked && request.BlockReason == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrBlockReasonRequired))
		return
	}

	// Clear blockReason if isBlocked is false
	if !request.IsBlocked {
		request.BlockReason = ""
	}

	customer, err := c.customerService.CreateCustomer(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, customer)
}