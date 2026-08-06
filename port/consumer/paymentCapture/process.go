package paymentCaptureConsumer

import (
	"context"
	"encoding/json"
	"fmt"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/monitoring"
	"github.com/paper-indonesia/pivot-backoffice/pkg/customMetric"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *paymentCaptureConsumer) PaymentCaptureProcess(ctx context.Context, body []byte, _ string) (err error) {
	ctx, span := otelTracer.Start(ctx, "port/consumer/paymentCapture/PaymentCaptureProcess")
	defer span.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from PaymentCaptureProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var (
		request          unifiedPaymentModel.ProcessCaptureRequest
		metricAttributes = map[string]any{}
	)
	if err := json.Unmarshal(body, &request); err != nil {
		return pkgErr.New(httpResponse.HttpErrUnprocessableContent, err)
	}

	defer func() {
		metricData := monitoring.CustomMetric{
			MetricInstrumentType: constant.MetricInstrumentTypeCounter,
			MetricValue:          1,
			ComponentName:        constant.ComponentNameUnifiedPayment,
			MetricName:           constant.MetricNameUnifiedPaymentCaptureProcessed,
			Attributes:           metricAttributes,
		}
		if err != nil {
			errType, errDetail := pkgErr.ExtractError(err)
			metricData.Attributes["errorType"] = errType
			metricData.Attributes["errorDetail"] = errDetail.Error()
		}
		errMetric := customMetric.RecordCustomMetric(ctx, &metricData)
		if errMetric != nil {
			c.logger.Error(ctx, "error when record processed payment capture custom metric", logger.Error(errMetric))
		}
	}()

	if err = c.unifiedPaymentSvc.ProcessCapture(ctx, &request); err != nil {
		return err
	}

	return nil
}
