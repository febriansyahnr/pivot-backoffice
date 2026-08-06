package walletTransaction

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("WalletTransactionService")

const (
	expirationForTransactionHistory = 15 * time.Minute
)

type service struct {
	log      logger.ILogger
	repo     repository.IWalletTransactionRepository
	cache    redisExt.IRedisExt
	storage  gcs.IGCSService
	internal port.IWalletTransactionInternalService
}

type Option func(*service)

func WithTestInternalService(svc port.IWalletTransactionInternalService) Option {
	return func(s *service) {
		s.internal = svc
	}
}

func New(log logger.ILogger, repo repository.IWalletTransactionRepository, cache redisExt.IRedisExt, storage gcs.IGCSService, opts ...Option) port.IWalletTransactionService {
	svc := &service{log, repo, cache, storage, nil}
	svc.internal = svc

	for _, opt := range opts {
		opt(svc)
	}
	return svc
}
