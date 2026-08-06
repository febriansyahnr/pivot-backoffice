package advanceairepository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/httpRequestExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AdvanceAiRepository")

type AdvanceAiRepository struct {
	config      *config.Config
	secret      *config.Secret
	logger      logger.ILogger
	httpRequest httpRequestExt.IHTTPRequest
}

type AdvanceAiRepositorFunc func(*AdvanceAiRepository)

var _ repository.IAmlProcessorRepository = (*AdvanceAiRepository)(nil)

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	httpRequest httpRequestExt.IHTTPRequest,
	depends ...AdvanceAiRepositorFunc,
) *AdvanceAiRepository {
	repos := &AdvanceAiRepository{
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

func (r *AdvanceAiRepository) baseURL() string {
	advanceAiFF := constant.GetAdvanceAiFeatureFlag(r.config.Environment)
	if advanceAiFF == nil {
		return r.config.AdvanceAIConfig.BaseURL
	}

	return advanceAiFF.BaseURL
}

func (r *AdvanceAiRepository) journeyID() string {
	advanceAiFF := constant.GetAdvanceAiFeatureFlag(r.config.Environment)
	if advanceAiFF == nil {
		return r.config.AdvanceAIConfig.JourneyID
	}

	return advanceAiFF.JourneyID
}
