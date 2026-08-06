package adjustment

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AdjustmentService")

type adjustmentService struct {
	cfg                   config.SlackConfig
	repo                  repository.IAdjustmentRepository
	gcs                   gcs.IGCSService
	orchestrator          service.IOrchestratorService
	merchantTopUpCallback service.IMerchantTopUpCallbackService
	merchantRepo          repository.IMerchantRepository
	rabbitMqExt           rabbitMqExt.IRabbitMQExt
	logger                logger.ILogger
	accountRepo           repository.IAccountRepository
}

func New(cfg config.SlackConfig, repo repository.IAdjustmentRepository, merchantRepo repository.IMerchantRepository) service.IAdjustmentService {
	return &adjustmentService{cfg: cfg, repo: repo, merchantRepo: merchantRepo}
}

func WithGCSService(s service.IAdjustmentService, gcs gcs.IGCSService) {
	s.(*adjustmentService).gcs = gcs
}

func WithRabbitMQ(s service.IAdjustmentService, rabbit rabbitMqExt.IRabbitMQExt) {
	s.(*adjustmentService).rabbitMqExt = rabbit
}

func WithOrchestratorService(s service.IAdjustmentService, orchestrator service.IOrchestratorService) {
	s.(*adjustmentService).orchestrator = orchestrator
}

func WithLogger(s service.IAdjustmentService, logger logger.ILogger) {
	s.(*adjustmentService).logger = logger
}

func WithAccountRepository(s service.IAdjustmentService, accountRepo repository.IAccountRepository) {
	s.(*adjustmentService).accountRepo = accountRepo
}

func WithMerchantTopUpCallbackService(s service.IAdjustmentService, callbackSvc service.IMerchantTopUpCallbackService) {
	s.(*adjustmentService).merchantTopUpCallback = callbackSvc
}
