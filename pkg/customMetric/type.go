package customMetric

import (
	"context"
	"errors"

	"github.com/paper-indonesia/pdk/v2/otelExt"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/monitoring"
)

var otelMetric otelExt.IOtelExt

func SetOtelExt(ext otelExt.IOtelExt) {
	otelMetric = ext
}

func getMetricInstrumentType(metricInstrumentType string) string {
	switch metricInstrumentType {
	case constant.MetricInstrumentTypeCounter:
		return constant.MetricInstrumentTypeCounter
	case constant.MetricInstrumentTypeGauge:
		return constant.MetricInstrumentTypeGauge
	case constant.MetricInstrumentTypeHistogram:
		return constant.MetricInstrumentTypeHistogram
	default:
		return constant.MetricInstrumentTypeCounter
	}
}

func RecordCustomMetric(ctx context.Context, request *monitoring.CustomMetric) error {
	if otelMetric == nil {
		return errors.New("otel metric is not initialized")
	}
	if request == nil {
		return errors.New("request cannot be nil")
	}

	return otelMetric.SendCustomMetric(ctx, otelExt.CustomMonitoringRequest{
		MeterName:            request.ComponentName,
		MetricName:           request.MetricName,
		MetricInstrumentType: getMetricInstrumentType(request.MetricInstrumentType),
		MetricValue:          request.MetricValue,
		Attributes:           request.Attributes,
	})
}
