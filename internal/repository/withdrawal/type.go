package withdrawalRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

type withdrawalRepository struct {
	*basicsql.Properties
	db mySqlExt.IMySqlExt
}

var (
	tableName  = "withdrawals"
	otelTracer = otel.Tracer("WithdrawalRepository")
)

func New(db mySqlExt.IMySqlExt) repository.IWithdrawalRepository {
	return &withdrawalRepository{
		basicsql.NewBasicSQLProperties(db), db,
	}
}
