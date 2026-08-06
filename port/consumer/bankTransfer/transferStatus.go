package bankTransferConsumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/bankTransfer"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) UpdateTransferStatus(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/bankTransfer/UpdateTransferStatus")
	defer segment.End()

	var (
		payload = &model.BankTransferResponseData{
			ProcessorReference: constant.SnapCoreProcessor,
		}
		metricAttributes = map[string]any{}
		start            = time.Now().UTC()
	)
	if err = json.Unmarshal(body, payload); err != nil {
		return fmt.Errorf("json Unmarshal: %v", err)
	}
	defer func() {
		end := time.Since(start)
		h.logger.Info(ctx, "Invoke update transfer status", logger.Any("payload", payload), logger.String("duration", end.String()), logger.Int64("durationMs", end.Milliseconds()), logger.Bool("completed", err == nil))

		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNamePayout,
			MetricName:           constant.MetricNamePayoutProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgError.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			h.logger.Error(ctx, "error when record processed payout custom metric", logger.Error(errMetric))
		}

	}()

	if payload.Transaction, err = h.ledgerSvc.FindByID(ctx, payload.ExternalID); err != nil {
		if errors.Is(err, constant.ErrDataNotFound) {
			h.logger.Info(ctx, fmt.Sprintf("Transaction with ID %s not found. Transfer status update ignored", payload.ExternalID))
			return nil
		}
		return err
	}
	metricAttributes["type"] = payload.Transaction.Type
	metricAttributes["status"] = payload.Transaction.Status

	switch payload.Transaction.Type {
	default:
		h.logger.Info(ctx, fmt.Sprintf("Transaction with ID %s transaction type is not registered. Transfer status update ignored", payload.ExternalID), logger.String("transactionType", payload.Transaction.Type))

	case constant.TypeDisbursement, constant.TypeBulkDisbursement:
		if payload.Transaction.Reference == constant.TypePaymentFundedPayout {
			return h.cardFundedPayoutSvc.UpdateBankTransferStatus(ctx, payload)
		}
		return h.disbursementSvc.ProcessUpdateTransferStatus(ctx, payload)

	case constant.TypeRefund:
		return h.refundProcSvc.ProcessUpdateBankTransferStatus(ctx, payload)

	case constant.TypeWithdrawal:
		return h.withdrawalSvc.UpdateBankTransferStatus(ctx, payload)
	}

	return nil
}
