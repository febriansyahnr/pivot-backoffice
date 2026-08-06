package cancelMoneyFlowService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CancelMoneyFlowService")

type CancelMoneyFlowService struct {
	logger      logger.ILogger
	repo        repository.IAccountTransactionRepository
	accountSvc  service.IAccountService
	ledgerSvc   service.ILedgerService
	merchantSvc service.IMerchantService
	queues      rabbitMqExt.IRabbitMQExt
}

func New(
	logger logger.ILogger,
	repo repository.IAccountTransactionRepository,
	accountSvc service.IAccountService,
	ledgerSvc service.ILedgerService,
	merchantSvc service.IMerchantService,
	opts ...OptionFunc,
) service.ILedgerMoneyFlowService {
	service := &CancelMoneyFlowService{
		logger:      logger,
		repo:        repo,
		accountSvc:  accountSvc,
		ledgerSvc:   ledgerSvc,
		merchantSvc: merchantSvc,
	}
	for _, opt := range opts {
		opt(service)
	}
	return service
}

type OptionFunc func(*CancelMoneyFlowService)

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) OptionFunc {
	return func(s *CancelMoneyFlowService) {
		s.queues = rmq
	}
}
