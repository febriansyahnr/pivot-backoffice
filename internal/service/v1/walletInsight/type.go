package walletInsightService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("WalletInsightService")

type WalletInsightService struct {
	orchestratorSvc service.IOrchestratorService
	redisClient     redisExt.IRedisExt
	logger          logger.ILogger
}

func New(
	orchestratorSvc service.IOrchestratorService,
	logger logger.ILogger,
	redis redisExt.IRedisExt,
) service.IWalletInsightService {
	return &WalletInsightService{
		orchestratorSvc: orchestratorSvc,
		redisClient:     redis,
		logger:          logger,
	}
}
