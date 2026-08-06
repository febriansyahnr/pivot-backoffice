package v1CrmPaymentController

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *handler) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/GetList")
	defer segment.End()

	var (
		err error
	)

	request, err := h.parseFilterParam(r)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	result, err := h.paymentSvc.GetListForInternalDashboard(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (h *handler) parseFilterParam(r *http.Request) (paymentModel.GetListFilterRequest, error) {
	var (
		opt paymentModel.GetListFilterRequest
		err error
	)
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = &d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = &d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.UUID = r.URL.Query().Get("uuid")
	opt.Status = r.URL.Query().Get("status")
	opt.ReferenceID = r.URL.Query().Get("referenceId")
	opt.PaymentMethod = r.URL.Query().Get("paymentMethod")
	opt.MerchantID = r.URL.Query().Get("merchantId")

	return opt, nil
}
