package settlementConsumerController

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *consumerHandler) ProcessPaymentSettlement(ctx context.Context, body []byte, channel string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/settlement/ProcessSettlement")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "Panic recovery from ProcessPaymentSettlement", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var (
		now     = time.Now()
		payload = &settlementModel.ProcessSettlementRequest{
			Type: constant.SettlementTransaction,
		}
		metricAttributes = map[string]any{}
	)

	defer monitor.WriteAndSend(
		ctx, "process-payment-settlement", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("merchant_id:%s", payload.MerchantID),
			}
		},
	)

	if err = json.Unmarshal(body, payload); err != nil {
		return pkgErrors.New(response.HttpErrRequest, err)
	}

	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentSettlementProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			h.logger.Error(ctx, "error when record processed settlement custom metric", logger.Error(errMetric))
		}
	}()
	metricAttributes["settlementType"] = payload.Type
	metricAttributes["merchantID"] = payload.MerchantID

	err = h.settlementSvc.ProcessSettlement(ctx, payload)
	if err != nil {
		h.logger.Error(ctx, "Error ProcessSettlement", logger.Error(err))
		return err
	}

	return nil
}
