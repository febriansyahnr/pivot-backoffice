package withdrawalRepository

import (
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/basicsql"
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
