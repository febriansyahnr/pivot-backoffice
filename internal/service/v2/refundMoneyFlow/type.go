package refundMoneyFlowService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RefundMoneyFlowService")

type RefundMoneyFlowService struct {
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
	service := &RefundMoneyFlowService{
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

type OptionFunc func(*RefundMoneyFlowService)

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) OptionFunc {
	return func(s *RefundMoneyFlowService) {
		s.queues = rmq
	}
}
