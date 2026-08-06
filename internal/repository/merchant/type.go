package merchant

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const (
	merchantsTable    = "merchants"
	merchantFeeTable  = "merchant_fees"
	merchantAuthTable = "merchant_auths"
)

var otelTracer = otel.Tracer("MerchantRepository")

type Options func(*MerchantRepository)

type MerchantRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
	config *config.Config

	// jsonMarshalFunc allows injecting custom marshal behavior for testing
	jsonMarshalFunc func(v interface{}) ([]byte, error)
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger, options ...Options) repository.IMerchantRepository {
	r := &MerchantRepository{
		Properties: basicsql.NewBasicSQLProperties(db),

		db:     db,
		logger: logger,
	}

	for _, opt := range options {
		opt(r)
	}
	return r
}

func WithServiceConfig(cfg *config.Config) Options {
	return func(r *MerchantRepository) {
		r.config = cfg
	}
}

// WithJSONMarshalFunc allows injecting custom JSON marshal function for testing
func WithJSONMarshalFunc(fn func(v interface{}) ([]byte, error)) Options {
	return func(r *MerchantRepository) {
		r.jsonMarshalFunc = fn
	}
}
