package vccSettlement

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("VccSettlementRepository")
	tableName  = "vcc_settlements"
)

type VccSettlementRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IVCCSettlementRepository {
	return &VccSettlementRepository{
		db:     db,
		logger: logger,
	}
}
