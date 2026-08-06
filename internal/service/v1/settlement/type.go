package settlementService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SettlementService")

type settlementService struct {
	logger logger.ILogger

	accountTransactionRepo repository.IAccountTransactionRepository

	paymentSvc          service.IPaymentService
	merchantSvc         service.IMerchantService
	internalSvc         service.ISettlementService
	cardFundedPayoutSvc service.ICardFundedPayoutService
}

type SettlementServiceFunc func(*settlementService)

func New(
	logger logger.ILogger,
	accountTransactionRepo repository.IAccountTransactionRepository,
	depends ...SettlementServiceFunc,
) service.ISettlementService {
	s := &settlementService{
		logger:                 logger,
		accountTransactionRepo: accountTransactionRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	s.internalSvc = s

	return s
}

func WithPaymentSvc(svc service.IPaymentService) SettlementServiceFunc {
	return func(s *settlementService) {
		s.paymentSvc = svc
	}
}

func WithMerchantSvc(svc service.IMerchantService) SettlementServiceFunc {
	return func(s *settlementService) {
		s.merchantSvc = svc
	}
}

func WithInternalSvc(svc service.ISettlementService, internalSvc service.ISettlementService) {
	svc.(*settlementService).internalSvc = internalSvc
}

func WithCardFundedPayoutSvc(svc service.ISettlementService, cardFundedPayoutSvc service.ICardFundedPayoutService) {
	svc.(*settlementService).cardFundedPayoutSvc = cardFundedPayoutSvc
}
