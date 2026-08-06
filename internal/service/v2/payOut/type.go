package payoutMoneyFlowService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PayoutMoneyFlowService")

type PayOutMoneyFlowService struct {
	logger      logger.ILogger
	repo        repository.IAccountTransactionRepository
	accountSvc  service.IAccountService
	ledgerSvc   service.ILedgerService
	merchantSvc service.IMerchantService
}

func New(
	logger logger.ILogger,
	repo repository.IAccountTransactionRepository,
	accountSvc service.IAccountService,
	ledgerSvc service.ILedgerService,
	merchantSvc service.IMerchantService,
) service.ILedgerMoneyFlowService {
	return &PayOutMoneyFlowService{
		logger:      logger,
		repo:        repo,
		accountSvc:  accountSvc,
		ledgerSvc:   ledgerSvc,
		merchantSvc: merchantSvc,
	}
}
