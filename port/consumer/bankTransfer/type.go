package bankTransferConsumer

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("BankTransferConsumer")

type handler struct {
	logger              logger.ILogger
	ledgerSvc           service.IOrchestratorService
	disbursementSvc     service.IDisbursementService
	refundProcSvc       service.IRefundProcessorService
	withdrawalSvc       service.IWithdrawalService
	cardFundedPayoutSvc service.ICardFundedPayoutService
	redisExt            redisExt.IRedisExt
}

type Service struct {
	LedgerSvc           service.IOrchestratorService
	DisbursementSvc     service.IDisbursementService
	RefundProcSvc       service.IRefundProcessorService
	WithdrawalSvc       service.IWithdrawalService
	CardFundedPayoutSvc service.ICardFundedPayoutService
	RedisExt            redisExt.IRedisExt
}

func New(logger logger.ILogger, service *Service) consumer.IBankTransferConsumer {
	return &handler{
		logger:              logger,
		ledgerSvc:           service.LedgerSvc,
		disbursementSvc:     service.DisbursementSvc,
		refundProcSvc:       service.RefundProcSvc,
		withdrawalSvc:       service.WithdrawalSvc,
		cardFundedPayoutSvc: service.CardFundedPayoutSvc,
		redisExt:            service.RedisExt,
	}
}
