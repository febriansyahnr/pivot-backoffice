package ledgerService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("LedgerService")

type LedgerService struct {
	logger          logger.ILogger
	repo            repository.IAccountTransactionRepository
	accountRepo     repository.IAccountRepository
	merchantSvc     service.IMerchantService
	customerSvc     service.ICustomerService
	accountSvc      service.IAccountService
	moneyFlowSvcMap map[string]service.ILedgerMoneyFlowService
	validatorSvc    service.ILedgerValidatorService
}

func New(
	logger logger.ILogger,
	repo repository.IAccountTransactionRepository,
	accountRepo repository.IAccountRepository,
	merchantSvc service.IMerchantService,
	customerSvc service.ICustomerService,
	accountSvc service.IAccountService,
) service.ILedgerService {
	svc := &LedgerService{
		logger:      logger,
		repo:        repo,
		accountRepo: accountRepo,
		merchantSvc: merchantSvc,
		customerSvc: customerSvc,
		accountSvc:  accountSvc,
	}
	svc.validatorSvc = svc
	return svc
}

func WithMoneyFlowService(s service.ILedgerService, transferType string, svc service.ILedgerMoneyFlowService) {
	ledgerService := s.(*LedgerService)
	if ledgerService.moneyFlowSvcMap == nil {
		ledgerService.moneyFlowSvcMap = make(map[string]service.ILedgerMoneyFlowService)
	}
	ledgerService.moneyFlowSvcMap[transferType] = svc
}

func WithValidatorService(s service.ILedgerService, svc service.ILedgerValidatorService) {
	s.(*LedgerService).validatorSvc = svc
}
