package walletTransaction

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (h *handler) mapQueryTransactionHistoryListToRequest(r *http.Request) (request model.MerchantTransactionHistoryListReq, err error) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/wallet/transaction/MerchantTransactionHistoryListRequest")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		return request, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
	}

	query := r.URL.Query()

	request = model.MerchantTransactionHistoryListReq{
		MerchantId:   user.MerchantId,
		StartDateReq: query.Get("startDate"), // Validation: iso_8601 (2006-01-02T15:04:05-07:00)
		EndDateReq:   query.Get("endDate"),   // Validation: iso_8601 (2006-01-02T15:04:05-07:00),gtecsfield=StartDateReq
		Type:         query.Get("type"),
		Status:       query.Get("status"),
		Id:           query.Get("id"),
		ReferenceId:  query.Get("referenceId"),
		Sort:         query.Get("sort"),
	}
	if request.Sort == "" {
		request.Sort = "-date"
	}
	request.Page, request.PerPage = httputil.GetPaginationQuery(r)

	if err = h.vld.StructCtx(ctx, &request); err != nil {
		return request, pkgErrs.New(response.HttpErrRequest, err)
	}

	request.StartDate = util.ParseISO8601DatetimeToUTC(request.StartDateReq)
	request.EndDate = util.ParseISO8601DatetimeToUTC(request.EndDateReq)

	if util.CalculateDaysBetween(request.StartDate, request.EndDate) > maxRangeDays {
		return request, pkgErrs.New(response.HttpErrRequest, fmt.Errorf(constant.ErrDateRangeLimitDaysFmt, maxRangeDays))

	} else if util.CalculateDaysBetween(request.StartDate, time.Now().UTC()) > maxBackdateDays {
		return request, pkgErrs.New(response.HttpErrRequest, fmt.Errorf(constant.ErrDateRangeLimitLastMonthsFmt, maxBackdateMonths))
	}

	return request, nil
}

func (h *handler) GetMerchantTransactionHistoryList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/wallet/transaction/GetMerchantTransactionHistoryList")
	defer segment.End()

	request, err := h.mapQueryTransactionHistoryListToRequest(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	resp, err := h.service.GetMerchantTransactionHistoryList(ctx, request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	response.SendApiResponsePaginationOK(w, resp.Data, resp.Meta)
}

func (h *handler) ExportMerchantTransactionHistoryList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/wallet/transaction/ExportMerchantTransactionHistoryList")
	defer segment.End()

	request, err := h.mapQueryTransactionHistoryListToRequest(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if resp, err := h.service.ExportMerchantTransactionHistoryList(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)
	} else {
		response.SendApiResponseOK(w, resp)
	}
}

func (h *handler) GetMerchantTransactionDetail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/wallet/transaction/GetMerchantTransactionDetail")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := r.PathValue("id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid transaction id format")))
		return
	}

	if resp, err := h.service.GetMerchantTransactionDetail(ctx, user.MerchantId, id); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
