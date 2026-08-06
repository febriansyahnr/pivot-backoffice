package fraudnetrepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("FraudNetRepository")

type FraudNetRepository struct {
	config      *config.Config
	secret      *config.Secret
	logger      logger.ILogger
	httpRequest httpRequestExt.IHTTPRequest
}

type FraudNetRepositoryFunc func(*FraudNetRepository)

var _ repository.IFdsProcessorRepository = (*FraudNetRepository)(nil)

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpRequest httpRequestExt.IHTTPRequest,
	depends ...FraudNetRepositoryFunc,
) *FraudNetRepository {
	repos := &FraudNetRepository{
		config:      config,
		secret:      secret,
		logger:      logger,
		httpRequest: httpRequest,
	}

	for _, fn := range depends {
		fn(repos)
	}

	return repos
}

func (r *FraudNetRepository) baseURL() string {
	fraudNetFF := constant.GetFraudNetFeatureFlag(r.config.Environment)
	if fraudNetFF == nil {
		return r.config.FraudNetConfig.BaseURL
	}

	return fraudNetFF.BaseURL
}
