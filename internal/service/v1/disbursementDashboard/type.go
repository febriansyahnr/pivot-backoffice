package disbursementDashboardService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DisbursementDashboardService")

type DisbursementDashboardService struct {
	logger                 logger.ILogger
	disbursementRepo       repository.IDisbursementRepository
	accountTransactionRepo repository.IAccountTransactionRepository
	accountRepo            repository.IAccountRepository
	orchestratorSvc        service.IOrchestratorService
}

func New(
	logger logger.ILogger,
	disbursementRepo repository.IDisbursementRepository,
	accountTransactionRepo repository.IAccountTransactionRepository,
	accountRepo repository.IAccountRepository,
	orchestratorSvc service.IOrchestratorService,
) service.IDisbursementDashboardService {
	return &DisbursementDashboardService{
		logger:                 logger,
		disbursementRepo:       disbursementRepo,
		accountTransactionRepo: accountTransactionRepo,
		accountRepo:            accountRepo,
		orchestratorSvc:        orchestratorSvc,
	}
}
