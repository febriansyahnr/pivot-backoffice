package refundRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	refundTable = "refunds"
)

var otelTracer = otel.Tracer("RefundRepository")

type RefundRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IRefundRepository {
	return &RefundRepository{
		Properties: basicsql.NewBasicSQLProperties(db),

		db:     db,
		logger: logger,
	}
}
