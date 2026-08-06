package transferService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type TransferServiceOption func(*TransferService)

type TransferService struct {
	logger         logger.ILogger
	ledgerSvc      service.ILedgerService
	accountSvc     service.IAccountService
	platformFeeSvc service.IPlatformFeeService
	merchantSvc    service.IMerchantService
	repo           repository.ITransferRepository
	paymentRepo    repository.IPaymentRepository
}

var otelTracer = otel.Tracer("TransferService")

func New(
	logger logger.ILogger,
	ledgerSvc service.ILedgerService,
	accountSvc service.IAccountService,
	platformFeeSvc service.IPlatformFeeService,
	merchantSvc service.IMerchantService,
	repo repository.ITransferRepository,
	opts ...TransferServiceOption,
) service.ITransferService {
	s := &TransferService{
		logger:         logger,
		ledgerSvc:      ledgerSvc,
		accountSvc:     accountSvc,
		platformFeeSvc: platformFeeSvc,
		merchantSvc:    merchantSvc,
		repo:           repo,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithPaymentRepository(paymentRepo repository.IPaymentRepository) TransferServiceOption {
	return func(s *TransferService) {
		s.paymentRepo = paymentRepo
	}
}
