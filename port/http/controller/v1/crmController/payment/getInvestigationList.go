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

func (h *handler) GetInvestigationList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payments/GetInvestigationList")
	defer segment.End()

	request, err := h.parseInvestigationFilterParam(r)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	result, err := h.paymentSvc.GetInvestigatedPayments(ctx, &request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

func (h *handler) parseInvestigationFilterParam(r *http.Request) (paymentModel.GetInvestigatedPaymentsFilterRequest, error) {
	var (
		opt paymentModel.GetInvestigatedPaymentsFilterRequest
		err error
	)

	opt.Page = 1
	opt.Limit = 20
	opt.Sort = "DESC"
	opt.SortBy = "updated_at"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("limit") != "" {
		opt.Limit, err = strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid limit format. Use number format instead"))
		}
		if opt.Limit > 100 {
			opt.Limit = 100
		}
	}

	if r.URL.Query().Get("fromDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("fromDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid fromDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.FromDate = &d
	}

	if r.URL.Query().Get("toDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("toDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, errors.New("invalid toDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.ToDate = &d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.InvestigationStatus = r.URL.Query().Get("investigationStatus")
	opt.PaymentReferenceID = r.URL.Query().Get("paymentReferenceId")
	opt.MerchantID = r.URL.Query().Get("merchantId")

	return opt, nil
}
