package customer

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMCustomerController) GetCustomerList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/customer/GetCustomerList")
	defer segment.End()

	merchantID := r.URL.Query().Get("merchant_id")
	if merchantID == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrMerchantIDRequired))
		return
	}

	pageStr := r.URL.Query().Get("page")
	perPageStr := r.URL.Query().Get("per_page")
	phoneNumber := r.URL.Query().Get("phone_number")

	page := int64(1)
	perPage := int64(10)

	if pageStr != "" {
		if parsedPage, err := strconv.ParseInt(pageStr, 10, 64); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if perPageStr != "" {
		if parsedPerPage, err := strconv.ParseInt(perPageStr, 10, 64); err == nil && parsedPerPage > 0 && parsedPerPage <= 100 {
			perPage = parsedPerPage
		}
	}

	customers, err := c.customerService.GetCustomerList(ctx, merchantID, phoneNumber, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, customers)
}