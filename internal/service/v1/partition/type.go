package partitionService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PartitionService")

type service struct {
	repository repository.ITablePartitionRepository
}

func New(repo repository.ITablePartitionRepository) port.ITablePartitionService {
	return &service{repo}
}
