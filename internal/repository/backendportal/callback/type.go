package callbackRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CallbackRepository")

type CallbackRepository struct {
	*basicsql.Properties
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.ICallbackRepository {
	return &CallbackRepository{
		Properties: basicsql.NewBasicSQLProperties(db),
		db:         db,
		logger:     logger,
	}
}

const (
	TableCallback    = "callbacks"
	TableCallbackLog = "callback_logs"
)
