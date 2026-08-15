package merchantRcn

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	merchantRCNSTable = "merchant_rcns"
)

var otelTracer = otel.Tracer("MerchantRcnRepository")

type Options func(*MerchantRcnRepository)

type MerchantRcnRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
	config *config.Config
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger, options ...Options) repository.IMerchantRcnRepository {
	r := &MerchantRcnRepository{

		db:     db,
		logger: logger,
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}
