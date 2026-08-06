package settlementHoldService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SettlementHoldService")

type settlementHoldService struct {
	logger logger.ILogger

	repo                   repository.ISettlementHoldRepository
	accountTransactionRepo repository.IAccountTransactionRepository

	paymentSvc    service.IPaymentService
	settlementSvc service.ISettlementService
}

func New(
	logger logger.ILogger,
	repo repository.ISettlementHoldRepository,
	paymentSvc service.IPaymentService,
	settlementSvc service.ISettlementService,
	accountTransactionRepo repository.IAccountTransactionRepository,
) service.ISettlementHoldService {
	s := &settlementHoldService{
		logger:                 logger,
		repo:                   repo,
		accountTransactionRepo: accountTransactionRepo,
		settlementSvc:          settlementSvc,
		paymentSvc:             paymentSvc,
	}

	return s
}
