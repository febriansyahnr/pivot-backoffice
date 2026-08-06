package v2InternalUnifiedPaymentController

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/GetList")
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

	request, err := c.parseFilterParam(r)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	request.MerchantID = merchantID

	if resp, err = c.unifiedPaymentSvc.GetSessionList(ctx, &request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (c *paymentController) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/GetByID")
	defer segment.End()

	var (
		err  error
		resp *unifiedPaymentModel.UnifiedPaymentSessionResponse
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

	id := chi.URLParam(r, "uuid")
	if err = uuid.Validate(id); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	request := &unifiedPaymentModel.GetUnifiedPaymentSessionRequest{
		PaymentSessionID: id,
		MerchantID:       merchantID,
	}
	if resp, err = c.unifiedPaymentSvc.GetSessionDetail(ctx, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}

func (c *paymentController) GetBinDetailByBinNumber(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/GetBinDetailByBinNumber")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	binNumber := r.PathValue("binNumber")
	if err := c.validate.VarCtx(ctx, binNumber, "required,numeric,min=6,max=8"); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, constant.NewErrStringRequest(response.HttpErrRequest, constant.ErrCodeV2APIValidationError, "Invalid BIN format"))
		return
	}

	request := unifiedPaymentModel.GetBinDetailRequest{
		MerchantId: merchantAuth.MerchantId,
		BinNumber:  binNumber,
	}
	result, err := c.unifiedPaymentSvc.GetCardBinDetail(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponseOK(w, result)
}

func (c *paymentController) parseFilterParam(r *http.Request) (paymentModel.GetListFilterRequest, error) {
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
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = &d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
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
	opt.ReferenceID = r.URL.Query().Get("clientReferenceId")
	opt.PaymentMethod = r.URL.Query().Get("paymentMethodType")

	return opt, nil
}
