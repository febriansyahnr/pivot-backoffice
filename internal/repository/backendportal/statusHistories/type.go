package statusHistoriesRepository

import (
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("StatusHistoriesRepository")

type statusHistoriesRepo struct {
	*basicsql.Properties

	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) repository.IStatusHistoriesRepository {
	return &statusHistoriesRepo{basicsql.NewBasicSQLProperties(db), db}
}
