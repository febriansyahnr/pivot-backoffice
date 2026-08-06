package creditcardCoreProcessorRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CreditcardCoreProcessorRepository")

type creditcardCoreProcessorRepository struct {
	config         *config.Config
	secret         *config.Secret
	logger         logger.ILogger
	httpRequest    httpRequestExt.IHTTPRequest
	cryptoProvider encryption.CryptoProvider
}

type Option func(r *creditcardCoreProcessorRepository)

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpRequest httpRequestExt.IHTTPRequest,
	options ...Option,
) repository.ICreditcardCoreProcessorRepository {
	repo := &creditcardCoreProcessorRepository{
		config:         config,
		secret:         secret,
		logger:         logger,
		httpRequest:    httpRequest,
		cryptoProvider: encryption.NewCryptoProvider(), // default
	}
	for _, opt := range options {
		opt(repo)
	}
	return repo
}

func WithCryptoProvider(provider encryption.CryptoProvider) Option {
	return func(r *creditcardCoreProcessorRepository) {
		r.cryptoProvider = provider
	}
}
