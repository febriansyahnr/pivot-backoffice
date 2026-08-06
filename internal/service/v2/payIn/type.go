package payInMoneyFlowService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PayInMoneyFlowService")

type PayInMoneyFlowService struct {
	logger      logger.ILogger
	repo        repository.IAccountTransactionRepository
	accountSvc  service.IAccountService
	merchantSvc service.IMerchantService
}

func New(
	logger logger.ILogger,
	repo repository.IAccountTransactionRepository,
	accountSvc service.IAccountService,
	merchantSvc service.IMerchantService,
) service.ILedgerMoneyFlowService {
	return &PayInMoneyFlowService{
		logger:      logger,
		repo:        repo,
		accountSvc:  accountSvc,
		merchantSvc: merchantSvc,
	}
}
