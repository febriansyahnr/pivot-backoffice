package v2InternalUnifiedPaymentController

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *paymentController) GetChargeList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/GetChargeList")
	defer segment.End()

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

	request, err := c.parseChargeFilterParam(r)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	request.MerchantID = merchantID

	if err := c.validateChargeFilterParam(r, request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	resp, err := c.unifiedPaymentSvc.GetChargeList(ctx, &request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (c *paymentController) GetChargeByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/internalController/unifiedPayment/GetChargeByID")
	defer segment.End()

	var (
		err  error
		resp *unifiedPaymentModel.ChargeResponse
	)

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	chargeID := chi.URLParam(r, "uuid")
	if err = uuid.Validate(chargeID); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchantAuth.MerchantId)
	}

	request := unifiedPaymentModel.GetUnifiedPaymentChargeRequest{
		ChargeID:   chargeID,
		MerchantID: merchantID,
	}

	if resp, err = c.unifiedPaymentSvc.GetChargeDetail(ctx, &request); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	resp.ExpiredAt = nil

	response.SendOpenApiResponseOK(w, resp)
}

func (c *paymentController) parseChargeFilterParam(r *http.Request) (opt unifiedPaymentModel.FilterChargeRequest, err error) {
	opt.Page = 1
	opt.PerPage = 10
	opt.Sort = "ASC"
	opt.SortBy = "createdAt"

	query := r.URL.Query()

	if query.Get("page") != "" {
		opt.Page, err = strconv.Atoi(query.Get("page"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid page format. Use number format instead"))
		}
	}

	if query.Get("perPage") != "" {
		opt.PerPage, err = strconv.Atoi(query.Get("perPage"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid perPage format. Use number format instead"))
		}
		if c.config.AppConfig.MaxPaginationPerPage > 0 && opt.PerPage > c.config.AppConfig.MaxPaginationPerPage {
			opt.PerPage = c.config.AppConfig.MaxPaginationPerPage
		}
	}

	if query.Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("startDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartCreatedAt = d
	}

	if query.Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, query.Get("endDate"))
		if err != nil {
			return opt, pkgErrors.New(response.HttpErrRequest, errors.New("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndCreatedAt = d
	}

	if opt.StartCreatedAt.IsZero() && opt.EndCreatedAt.IsZero() {
		opt.EndCreatedAt = time.Now().UTC()
		opt.StartCreatedAt = opt.EndCreatedAt.AddDate(0, 0, -31)
	}

	if query.Get("sort") != "" {
		opt.Sort = query.Get("sort")
	}

	if query.Get("sortBy") != "" {
		opt.SortBy = query.Get("sortBy")
	}

	opt.UUID = query.Get("id")
	opt.Status = query.Get("status")
	opt.ClientReferenceID = query.Get("clientReferenceId")
	opt.PaymentSessionID = query.Get("paymentSessionId")

	return opt, nil
}

func (c *paymentController) validateChargeFilterParam(r *http.Request, params unifiedPaymentModel.FilterChargeRequest) error {
	if params.StartCreatedAt.IsZero() || params.EndCreatedAt.IsZero() {
		return pkgErrors.New(response.HttpErrRequest, errors.New("startDate or endDate cannot be empty"))
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	if err := c.validate.Var(params.Sort, "omitempty,oneof=ASC DESC"); err != nil {
		return pkgErrors.New(response.HttpErrRequest, errors.New("sort parameter must be either ASC or DESC"))
	}
	return nil
}
