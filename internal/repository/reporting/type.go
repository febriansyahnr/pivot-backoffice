package reportingRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ReportingRepository")

type repository struct {
	db        mySqlExt.IMySqlExt
	logger    logger.ILogger
	appConfig config.AppConfig
}

func New(db mySqlExt.IMySqlExt, log logger.ILogger, appConfig config.AppConfig) port.IReportingRepository {
	return &repository{db, log, appConfig}
}

func (repository) TableName() string {
	return "report_balance_histories"
}
