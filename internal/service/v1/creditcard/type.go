package creditcard

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CreditCardService")

const serviceName = "Creditcard-Service"

type CreditCardService struct {
	config      *config.Config
	logger      logger.ILogger
	rabbitMqExt rabbitMqExt.IRabbitMQExt
	redis       redisExt.IRedisExt

	// repository
	paymentRepo                 repository.IPaymentRepository
	paymentMethodRepo           repository.IPaymentMethodRepository
	creditcardCoreProcessorRepo repository.ICreditcardCoreProcessorRepository
	customerRepo                repository.ICustomerRepository
	merchantRepo                repository.IMerchantRepository
	accountTransactionRepo      repository.IAccountTransactionRepository

	// service
	feeSvc           service.IFeeService
	orchestratorSvc  service.IOrchestratorService
	paymentLedgerSvc service.IPaymentLedgerService
	paymentMethodSvc service.IPaymentMethodService
	paymentSvc       service.IPaymentService
}

type CreditCardServiceFunc func(*CreditCardService)

func New(
	config *config.Config,
	logger logger.ILogger,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	paymentRepo repository.IPaymentRepository,
	paymentMethodRepo repository.IPaymentMethodRepository,
	creditcardCoreProcessorRepo repository.ICreditcardCoreProcessorRepository,
	depends ...CreditCardServiceFunc,
) service.ICreditCardService {
	d := &CreditCardService{
		config:                      config,
		logger:                      logger,
		rabbitMqExt:                 rabbitMqExt,
		paymentRepo:                 paymentRepo,
		paymentMethodRepo:           paymentMethodRepo,
		creditcardCoreProcessorRepo: creditcardCoreProcessorRepo,
	}

	for _, fn := range depends {
		fn(d)
	}

	return d
}

func WithFeeService(svc service.IFeeService) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.feeSvc = svc
	}
}

func WithPaymentLedgerService(svc service.IPaymentLedgerService) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.paymentLedgerSvc = svc
	}
}

func WithOrchestratorService(svc service.IOrchestratorService) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.orchestratorSvc = svc
	}
}

func WithCustomerRepo(repo repository.ICustomerRepository) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.customerRepo = repo
	}
}

func WithMerchantRepo(repo repository.IMerchantRepository) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.merchantRepo = repo
	}
}

func WithAccountTransactionRepo(repo repository.IAccountTransactionRepository) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.accountTransactionRepo = repo
	}
}

func WithPaymentMethodService(svc service.IPaymentMethodService) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.paymentMethodSvc = svc
	}
}

func WithRedis(redis redisExt.IRedisExt) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.redis = redis
	}
}

func WithPaymentService(svc service.IPaymentService) CreditCardServiceFunc {
	return func(ds *CreditCardService) {
		ds.paymentSvc = svc
	}
}
