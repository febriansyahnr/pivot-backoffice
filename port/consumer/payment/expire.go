package paymentConsumerController

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *paymentConsumer) ProcessPaymentExpiration(ctx context.Context, body []byte, channel string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/payment/ProcessPaymentExpiration")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from ProcessPaymentExpiration", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	request := paymentModel.ExpiringPayment{}

	now := time.Now()
	defer monitor.WriteAndSend(
		ctx, "process-payment-expiration", now, nil, err, func() []string {
			return []string{
				fmt.Sprintf("merchantId:%s", request.MerchantID),
			}
		},
	)

	if err = json.Unmarshal(body, &request); err != nil {
		return pkgError.New(response.HttpErrUnprocessableContent, err)
	}
	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentExpired,
			Attributes: map[string]any{
				"merchantId": request.MerchantID,
				"status":     request.ChargeStatus,
			},
		}
		if err != nil {
			errType, errDetail := pkgError.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "error when record expired payment custom metric", logger.Error(errMetric))
		}
	}()

	return c.paymentSvc.ExpirePayment(ctx, request)
}
