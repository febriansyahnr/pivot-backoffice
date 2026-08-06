package paymentDatamart

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/datamart"
	"github.com/paper-indonesia/pivot-backoffice/pkg/bigquery"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type repository struct {
	config config.BigQueryConfig
	db     bigquery.IBigQueryService
	logger logger.ILogger
}

var otelTracer = otel.Tracer("PaymentDatamartRepository")

func New(cfg config.BigQueryConfig, db bigquery.IBigQueryService, log logger.ILogger) datamart.IDatamartPaymentMetrics {
	return &repository{
		config: cfg, db: db, logger: log,
	}
}
