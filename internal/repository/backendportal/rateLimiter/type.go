package rateLimiter

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer             = otel.Tracer("RateLimiterRepository")
	MerchantRateLimitTable = "rate_limit_configurations"
)

type RateLimiterRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IRateLimiterRepository {
	return &RateLimiterRepository{
		db:     db,
		logger: logger,
	}
}
