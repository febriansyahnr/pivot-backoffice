package paymentRepository

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PaymentRepository")

const tableName = "payments"

// ConsulRetriever interface for testing purposes
type ConsulRetriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}

type Option func(*PaymentRepository)
type PaymentRepository struct {
	config          *config.Config
	secret          *config.Secret
	db              mySqlExt.IMySqlExt
	logger          logger.ILogger
	appConfig       *config.AppConfig
	consulRetriever ConsulRetriever // For testing purposes only
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger, options ...Option) repository.IPaymentRepository {
	r := &PaymentRepository{
		db:     db,
		logger: logger,
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func (r *PaymentRepository) WithConfig(config *config.Config) {
	r.config = config
}

func (r *PaymentRepository) WithSecret(secret *config.Secret) {
	r.secret = secret
}

func WithAppConfig(config *config.AppConfig) Option {
	return func(dr *PaymentRepository) {
		dr.appConfig = config
	}
}

// WithConsulRetriever allows injecting a mock retriever for testing purposes
func WithConsulRetriever(retriever ConsulRetriever) Option {
	return func(r *PaymentRepository) {
		r.consulRetriever = retriever
	}
}
