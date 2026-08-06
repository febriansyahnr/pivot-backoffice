package merchantForbiddenUsecase

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	tableName = "merchant_forbidden_usecases"
)

var otelTracer = otel.Tracer("MerchantForbiddenUsecaseRepository")

type MerchantForbiddenUsecaseRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IMerchantForbiddenUsecaseRepository {
	return &MerchantForbiddenUsecaseRepository{
		db:     db,
		logger: logger,
	}
}
