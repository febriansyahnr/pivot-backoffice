package partition

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/tablePartitionExt"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PartitionCronHandler")

type tablePartitionHandler struct {
	logger           logger.ILogger
	tablePartition   tablePartitionExt.IPartitionTable
	partitionService service.ITablePartitionService
}

func New(
	logger logger.ILogger,
	tablePartition tablePartitionExt.IPartitionTable,
	partitionService service.ITablePartitionService,
) handler.ITablePartitionHandler {
	return &tablePartitionHandler{
		logger:           logger,
		tablePartition:   tablePartition,
		partitionService: partitionService,
	}
}
