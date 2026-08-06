package orchestratorController

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	orchestratorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *OrchestratorController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "http/controller/v1/orchestrator/GetList")
	defer segment.End()

	filter, err := c.getQueryForTransactionHistory(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	page := int64(1)

	// Validation and parsing
	pageStr := r.URL.Query().Get("page")
	if pageStr != "" {
		page, err = strconv.ParseInt(pageStr, 10, 64)
		if err != nil {
			response.SendApiResponseError(ctx, w, pkgErrs.New(
				response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead")))
			return
		}
	}

	var perPage int64 = constant.DefaultPaginationPageSize
	if c.config != nil {
		perPage = c.config.AppConfig.PaginationPerPage
	}

	list, err := &commonModel.PaginationResponse{}, error(nil)
	if constant.IsEnableBalanceHistoryViaDataReporting(filter.MerchantID) {
		sort := strings.TrimSpace(r.URL.Query().Get("sort"))
		if sort == "" {
			sort = "-date"
		}
		filter.FilteredSortQuery, _ = sortFieldForGetListViaDataReporting(sort)

		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginReporting)
		list, err = c.reportingSvc.ListBalanceHistory(ctx, filter, page, perPage)
	} else {

		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginRaw)
		list, err = c.orchestratorSvc.GetList(ctx, filter, page, perPage)
	}
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrInternal, err))
		return
	}
	response.SendApiResponsePaginationOK(w, list.Data, list.Meta)
}

func (h *OrchestratorController) GetOpenApiBalanceHistories(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/orchestrator/GetBalanceHistories")
	defer segment.End()

	claims, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	err, query := error(nil), r.URL.Query()
	ctx = context.WithValue(ctx, constant.CtxFeatureName, constant.FeatureBalanceHistoryOpenApi)

	// Setting query parameter values
	filter := &orchestratorModel.TransactionHistoryFilterRequest{
		MerchantID:    claims.MerchantId,
		Status:        query.Get("status"),
		TransactionId: query.Get("transactionId"),
	}
	if trxType := query.Get("transactionType"); trxType != "" {
		filter.TrxTypes = []string{trxType}
	}
	if balanceType := query.Get("accountType"); balanceType != "" {
		filter.BalanceTypes = []string{balanceType}
	}

	if filter.FilteredSortQuery = query.Get("sort"); filter.FilteredSortQuery == "" {
		filter.FilteredSortQuery = "-date"
	}
	if filter.FilteredSortQuery, ok = sortFieldForGetList(filter.FilteredSortQuery); !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidSortOrder))
		return
	}

	if filter.StartSettlementDate, err = time.ParseInLocation(constant.DatetimeRFC3339Format, query.Get("startDate"), time.UTC); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidStartDateFmt))
		return
	}
	if filter.EndSettlementDate, err = time.ParseInLocation(constant.DatetimeRFC3339Format, query.Get("endDate"), time.UTC); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidEndDateFmt))
		return
	}
	if filter.StartSettlementDate.After(filter.EndSettlementDate) {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrFilterDateInput))
		return

	} else if (filter.EndSettlementDate.Sub(filter.StartSettlementDate).Hours()/24) > maxDateRangeInDays || (time.Since(filter.StartSettlementDate).Hours()/24) > maxBackdateInDays {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrDateRangeExceedLimit))
		return
	}

	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		// Validation of sub merchant status is already in the middleware and if the sub merchant is not found, it will be handled directly there.
		if merchant, err := h.merchantSvc.FindMerchantByID(ctx, subMerchantId); err != nil || merchant.ParentID.String != claims.MerchantId {
			if err == nil {
				err = pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
			}
			response.SendOpenApiNonSnapResponseError(ctx, w, err)
			return
		}
		filter.MerchantID = subMerchantId
	}

	page := int64(constant.DefaultPage)
	perPage := h.config.AppConfig.PaginationPerPage

	if strPage := query.Get("page"); strPage != "" {
		result, err := strconv.ParseInt(strPage, 10, 64)
		if err != nil || result < minPaginationValue {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
		page = result
	}

	if strPerPage := query.Get("perPage"); strPerPage != "" {
		result, err := strconv.ParseInt(strPerPage, 10, 64)
		if err != nil || result < minPaginationValue || result > maxPaginationPage {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
		perPage = result
	}

	var history *commonModel.PaginationResponse
	if constant.IsEnableBalanceHistoryViaDataReporting(filter.MerchantID) {
		sort := strings.TrimSpace(r.URL.Query().Get("sort"))
		if sort == "" {
			sort = "-date"
		}
		filter.FilteredSortQuery, _ = sortFieldForGetListViaDataReporting(sort)

		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginReporting)
		history, err = h.reportingSvc.ListBalanceHistory(ctx, filter, page, perPage)
	} else {

		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginRaw)
		history, err = h.orchestratorSvc.GetList(ctx, filter, page, perPage)
	}
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	response.SendOpenApiResponsePaginationOK(w, history.Data, history.Meta)
}
