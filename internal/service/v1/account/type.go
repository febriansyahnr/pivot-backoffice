package accountService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AccountService")

type account struct {
	logger                      logger.ILogger
	accountTransactionRepo      repository.IAccountTransactionRepository
	accountRepo                 repository.IAccountRepository
	dailyAccountTransactionRepo repository.IDailyAccountTransactionRepository
	customerSvc                 service.ICustomerService
	merchantSvc                 service.IMerchantService
}

func New(
	logger logger.ILogger,
	accountTransactionRepo repository.IAccountTransactionRepository,
	accountRepo repository.IAccountRepository,
	dailyAccountTransactionRepo repository.IDailyAccountTransactionRepository,
) service.IAccountService {
	return &account{
		logger:                      logger,
		accountTransactionRepo:      accountTransactionRepo,
		accountRepo:                 accountRepo,
		dailyAccountTransactionRepo: dailyAccountTransactionRepo,
	}
}

func WithCustomerService(service service.IAccountService, customerService service.ICustomerService) {
	service.(*account).customerSvc = customerService
}

func WithMerchantService(service service.IAccountService, merchantService service.IMerchantService) {
	service.(*account).merchantSvc = merchantService
}
