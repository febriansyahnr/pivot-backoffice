package reconciliation

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tableName = "reconciliations"
var otelTracer = otel.Tracer(tableName)

type ReconciliationRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) *ReconciliationRepository {
	return &ReconciliationRepository{
		Properties: basicsql.NewBasicSQLProperties(db),
		db:         db,
		logger:     logger,
	}
}

var _ repository.IReconciliationRepository = (*ReconciliationRepository)(nil)
