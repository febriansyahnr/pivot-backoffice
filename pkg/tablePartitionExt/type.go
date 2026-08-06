package tablePartitionExt

import (
	"context"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PartitionCronHandler")

type IPartitionTable interface {
	CreateDayRangePartition(ctx context.Context, cfg PartitionConfig) error
}

type partitionTable struct {
	db          mySqlExt.IMySqlExt
	log         logger.ILogger
	tableSchema string
}

type PartitionConfig struct {
	TableName          string
	TotalPartition     int
	StartedAt          time.Time
	Parameter          string
	IsPreciseTimestamp bool
}

// Create New instance of an IPartitionTable.
func New(db mySqlExt.IMySqlExt, logger logger.ILogger, tableSchema string) IPartitionTable {
	partitionTable := &partitionTable{
		db:          db,
		log:         logger,
		tableSchema: tableSchema,
	}

	return partitionTable
}
