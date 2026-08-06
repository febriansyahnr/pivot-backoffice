package xbPayoutConsumerController

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pdk/v2/logger"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (h *consumerHandler) UpdateStatus(ctx context.Context, body []byte, channel string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/xbPayout/UpdateStatus")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "Panic recovery from XbPayoutUpdateStatus", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var (
		now              = time.Now()
		payload          = &xbModel.ConsumePayoutStatusChangeRequest{}
		metricAttributes = map[string]any{}
	)

	defer monitor.WriteAndSend(
		ctx, "xb-payout-status-change", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("acquirer_transaction_id:%s", payload.AcquirerTransactionId),
				fmt.Sprintf("status:%s", payload.Status),
				fmt.Sprintf("timestamp:%s", payload.Timestamp.Format(time.RFC3339)),
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
			ComponentName:        constant.ComponentNameXB,
			MetricName:           constant.MetricNameXBUpdateStatus,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgErrors.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			h.logger.Error(ctx, "error when record processed xb payout custom metric", logger.Error(errMetric))
		}
	}()

	metricAttributes["status"] = payload.Status

	return h.service.UpdateStatusFromProcessor(ctx, payload)
}
