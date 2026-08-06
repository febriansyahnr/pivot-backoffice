package feeService

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("FeeService")
var tz, _ = time.LoadLocation(constant.TimeLoc)

type FeeService struct {
	logger                 logger.ILogger
	feeRepo                repository.IFeeRepository
	merchantRepo           repository.IMerchantRepository
	accountTransactionRepo repository.IAccountTransactionRepository

	// service
	paymentMethodSvc service.IPaymentMethodService
	orchestratorSvc  service.IOrchestratorService
	redis            redisExt.IRedisExt
	config           *config.Config
}

type FeeServiceFunc func(*FeeService)

func New(logger logger.ILogger,
	feeRepo repository.IFeeRepository,
	merchantRepo repository.IMerchantRepository,
	depends ...FeeServiceFunc,
) service.IFeeService {
	s := &FeeService{
		logger:       logger,
		feeRepo:      feeRepo,
		merchantRepo: merchantRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithPaymentMethodService(svc service.IPaymentMethodService) FeeServiceFunc {
	return func(ds *FeeService) {
		ds.paymentMethodSvc = svc
	}
}

func WithAccountTransactionRepository(repo repository.IAccountTransactionRepository) FeeServiceFunc {
	return func(s *FeeService) {
		s.accountTransactionRepo = repo
	}
}

func WithOrchestratorService(svc service.IOrchestratorService) FeeServiceFunc {
	return func(s *FeeService) {
		s.orchestratorSvc = svc
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) FeeServiceFunc {
	return func(s *FeeService) {
		s.redis = rdb
	}
}

func WithConfig(config *config.Config) FeeServiceFunc {
	return func(s *FeeService) {
		s.config = config
	}
}
