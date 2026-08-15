package liveFeature

import (
	"context"

	"github.com/paper-indonesia/pdk/v2/goff/retriever"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	liveFeature "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/liveFeature"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

const tableName = "services"

var otelTracer = otel.Tracer("LiveFeatureRepository")

// ConsulRetriever is an interface for consul retriever operations
type ConsulRetriever interface {
	Retrieve(ctx context.Context) ([]byte, error)
}

// ConsulRetrieverFactoryFunc is a function type for creating consul retrievers
type ConsulRetrieverFactoryFunc func(consulAddr, key, token string) (ConsulRetriever, error)

type LiveFeatureRepository struct {
	config                 *config.Config
	secret                 *config.Secret
	db                     mySqlExt.IMySqlExt
	logger                 logger.ILogger
	currentVersion         liveFeature.AppVersion // Add currentVersion field
	consulRetrieverFactory ConsulRetrieverFactoryFunc
}

// defaultConsulRetrieverFactory wraps pdkRetriever.NewConsulRetriever
func defaultConsulRetrieverFactory(consulAddr, key, token string) (ConsulRetriever, error) {
	return retriever.NewConsulRetriever(consulAddr, key, token)
}

// New function to initialize the repository
func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.ILiveFeatureRepository {
	return &LiveFeatureRepository{
		db:                     db,
		logger:                 logger,
		currentVersion:         liveFeature.AppVersion{Versions: make(map[string]string)}, // Initialize currentVersion
		consulRetrieverFactory: defaultConsulRetrieverFactory,                             // Default to real implementation
	}
}

// WithConfig allows injecting the config into the repository
func (r *LiveFeatureRepository) WithConfig(config *config.Config) {
	r.config = config
}

// WithSecret allows injecting the secret into the repository
func (r *LiveFeatureRepository) WithSecret(secret *config.Secret) {
	r.secret = secret
}
