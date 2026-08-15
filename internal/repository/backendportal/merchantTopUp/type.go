package merchantTopUp

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const tableName = "merchant_top_up_references"

var otelTracer = otel.Tracer("MerchantTopUpRepository")

type Option func(*merchantTopUpRepository)

type merchantTopUpRepository struct {
	db        mySqlExt.IMySqlExt
	logger    logger.ILogger
	appConfig *config.AppConfig
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger, options ...Option) repository.IMerchantTopUpRepository {
	r := &merchantTopUpRepository{
		db:     db,
		logger: logger,
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func WithAppConfig(cfg *config.AppConfig) Option {
	return func(r *merchantTopUpRepository) {
		r.appConfig = cfg
	}
}
