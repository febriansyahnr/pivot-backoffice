package adjustment

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AdjustmentRepository")

type adjustment struct {
	*basicsql.Properties

	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) repository.IAdjustmentRepository {
	return &adjustment{basicsql.NewBasicSQLProperties(db), db}
}
