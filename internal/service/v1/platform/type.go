package platformService

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PlatformService")

type PlatformService struct {
	logger                logger.ILogger
	disbursementService   service.IDisbursementService
	paymentService        service.IPaymentService
	merchantService       service.IMerchantService
	transferService       service.ITransferService
	userService           service.IUserService
	withdrawalService     service.IWithdrawalService
	merchantTopUpService  service.IMerchantTopUpService
	orchestratorService   service.IOrchestratorService
	unifiedPaymentService service.IUnifiedPaymentService
}

type PlatformServiceFunc func(*PlatformService)

func New(
	logger logger.ILogger,
	disbursementService service.IDisbursementService,
	paymentService service.IPaymentService,
	merchantService service.IMerchantService,
	transferService service.ITransferService,
	withdrawalService service.IWithdrawalService,
	merchantTopUpService service.IMerchantTopUpService,
	opts ...PlatformServiceFunc,
) service.IPlatformService {
	s := &PlatformService{
		logger:               logger,
		disbursementService:  disbursementService,
		paymentService:       paymentService,
		merchantService:      merchantService,
		transferService:      transferService,
		withdrawalService:    withdrawalService,
		merchantTopUpService: merchantTopUpService,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func WithUnifiedPaymentService(service service.IUnifiedPaymentService) PlatformServiceFunc {
	return func(ps *PlatformService) {
		ps.unifiedPaymentService = service
	}
}

func WithUserService(service service.IUserService) PlatformServiceFunc {
	return func(ps *PlatformService) {
		ps.userService = service
	}
}

func WithOrchestratorService(service service.IOrchestratorService) PlatformServiceFunc {
	return func(ps *PlatformService) {
		ps.orchestratorService = service
	}
}
