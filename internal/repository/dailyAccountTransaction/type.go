package dailyAccountTransactionRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DailyAccountTransactionRepository")

type DailyAccountTransactionRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	db mySqlExt.IMySqlExt,
	logger logger.ILogger,
) repository.IDailyAccountTransactionRepository {
	return &DailyAccountTransactionRepository{
		Properties: basicsql.NewBasicSQLProperties(db),

		db:     db,
		logger: logger,
	}
}

const tableName = "daily_account_transactions"
