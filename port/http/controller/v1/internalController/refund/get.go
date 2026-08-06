package v1InternalRefundController

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *refundController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/GetChargeList")
	defer segment.End()

	var (
		err  error
		resp *commonModel.PaginationResponse
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)
	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	request, err := c.ParseRefundFilterParam(r)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	request.MerchantID = merchantID

	resp, err = c.refundSvc.GetRefundList(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (c *refundController) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/GetChargeList")
	defer segment.End()

	var (
		err     error
		request refundModel.FilterRefundRequest
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	request.MerchantID = merchantID
	request.UUID = r.PathValue("uuid")

	if request.UUID == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("id is required")))
		return
	}

	refund, err := c.refundSvc.GetRefundDetail(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, refund)
}

func (c *refundController) ParseRefundFilterParam(r *http.Request) (refundModel.FilterRefundRequest, error) {
	var (
		opt refundModel.FilterRefundRequest
		err error
	)
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	if r.URL.Query().Get("page") != "" {
		opt.Page, err = strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startCreatedAt") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startCreatedAt"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = &d
	}

	if r.URL.Query().Get("endCreatedAt") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endCreatedAt"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = &d
	}

	if r.URL.Query().Get("sort") != "" {
		opt.Sort = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.UUID = r.URL.Query().Get("id")
	opt.Status = r.URL.Query().Get("status")
	opt.ClientReferenceID = r.URL.Query().Get("clientReferenceId")
	opt.PaymentSessionID = r.URL.Query().Get("paymentSessionId")

	return opt, nil
}
