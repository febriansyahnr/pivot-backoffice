package orchestratorController

import (
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *OrchestratorController) ExportToExcelTransactionHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "http/controller/v1/orchestrator/ExportToExcelTransactionHistory")
	defer segment.End()

	filter, err := c.getQueryForTransactionHistory(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if constant.IsEnableBalanceHistoryViaDataReporting(filter.MerchantID) {
		sort := strings.TrimSpace(r.URL.Query().Get("sort"))
		if sort == "" {
			sort = "-date"
		}
		filter.FilteredSortQuery, _ = sortFieldForGetListViaDataReporting(sort)

		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginReporting)
	} else {
		w.Header().Set(constant.HeaderDataOrigin, constant.DataOriginRaw)
	}
	if err := c.orchestratorSvc.GenExcelForTransactionHistories(ctx, &model.FileWriter{Writer: w}, filter); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
