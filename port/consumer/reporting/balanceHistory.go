package reportingConsumer

import (
	"context"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	cdcModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cdc"
	reportingModel "github.com/paper-indonesia/pivot-backoffice/internal/model/reporting"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"

	"github.com/paper-indonesia/pdk/v2/logger"
)

// BalanceHistory processes CDC events from account_transactions table
// and upserts the data to balance history for reporting purposes.
func (h *handler) BalanceHistory(ctx context.Context, data []byte) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/reporting/BalanceHistory")
	defer segment.End()

	ctx = context.WithValue(ctx, pdkConst.CtxRequestSourceKey, "data-reporting-service")

	event, err := cdcModel.ParseEvent[cdcModel.AccountTransaction](data)
	if err != nil {
		h.logger.Error(ctx, "Failed to parse event from the account_transactions table", logger.Error(err), logger.ByteString("event", data))
		return pkgErrs.NewNonRetryableError(err)
	}

	if event.IsEmpty() {
		h.logger.Warn(ctx, "Missing before and after data in the event from the account_transactions table", logger.ByteString("event", data))
		return pkgErrs.NewNonRetryableError(constant.ErrMalformedRequestBodyPayload)
	}

	h.logger.Info(ctx, "Process balance history change events", logger.String("operation", event.Op), logger.Any("payload", event.GetCurrent()))

	return h.service.UpsertBalanceHistory(ctx, &reportingModel.UpsertBalanceHistoryRequest{Event: event})
}
