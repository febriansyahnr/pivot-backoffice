package partitionRepository

import (
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PartitionRepository")

type repository struct{ db mySqlExt.IMySqlExt }

func New(db mySqlExt.IMySqlExt) port.ITablePartitionRepository {
	return &repository{db}
}
