package settlementHold

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer       = otel.Tracer("SettlementHoldRepository")
	tableName        = "settlement_holds"
	historyTableName = "settlement_hold_histories"
)

type settlementHoldRepo struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.ISettlementHoldRepository {
	return &settlementHoldRepo{basicsql.NewBasicSQLProperties(db), db, logger}
}
