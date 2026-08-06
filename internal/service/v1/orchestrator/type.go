package orchestrator_service

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/tracer"

	"github.com/paper-indonesia/pdk/v2/logger"
)

var otelTracer = tracer.New("OrchestratorService")

type OrchestratorService struct {
	logger                 logger.ILogger
	gcs                    gcs.IGCSService
	redis                  redisExt.IRedisExt
	accountTransactionRepo repository.IAccountTransactionRepository
	accountRepo            repository.IAccountRepository
	reportingRepo          repository.IReportingRepository
	accountSvc             service.IAccountService
}

type OrchestratorServiceFunc func(*OrchestratorService)

func New(
	logger logger.ILogger,
	gcs gcs.IGCSService,
	accountTransactionRepo repository.IAccountTransactionRepository,
	accountRepo repository.IAccountRepository,
	depends ...OrchestratorServiceFunc,
) service.IOrchestratorService {
	s := &OrchestratorService{
		logger:                 logger,
		gcs:                    gcs,
		accountTransactionRepo: accountTransactionRepo,
		accountRepo:            accountRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithRedisClient(rdb redisExt.IRedisExt) OrchestratorServiceFunc {
	return func(ds *OrchestratorService) {
		ds.redis = rdb
	}
}

func WithAccountService(svc service.IOrchestratorService, accountSvc service.IAccountService) {
	svc.(*OrchestratorService).accountSvc = accountSvc
}

func WithReportingRepository(svc service.IOrchestratorService, reportingRepo repository.IReportingRepository) {
	svc.(*OrchestratorService).reportingRepo = reportingRepo
}
