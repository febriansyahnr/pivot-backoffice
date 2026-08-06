package beneficiaryAccountService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("BeneficiaryAccountService")

type BeneficiaryAccountServiceFunc func(*BeneficiaryAccountService)
type BeneficiaryAccountService struct {
	logger                 logger.ILogger
	routingProcessorSvc    service.IRoutingProcessorService
	beneficiaryAccountRepo repository.IBeneficiaryAccountRepository
	accountInquiries       repository.IAccountInquiriesRepository
	snapCore               repository.ISnapCoreRepository
	config                 *config.Config
}

func New(
	logger logger.ILogger,
	beneficiaryAccountRepo repository.IBeneficiaryAccountRepository,
	accountInquiriesRepo repository.IAccountInquiriesRepository,
	snapCore repository.ISnapCoreRepository,
	depends ...BeneficiaryAccountServiceFunc,
) service.IBeneficiaryAccountService {
	svc := &BeneficiaryAccountService{
		logger:                 logger,
		beneficiaryAccountRepo: beneficiaryAccountRepo,
		accountInquiries:       accountInquiriesRepo,
		snapCore:               snapCore,
	}

	for _, fn := range depends {
		fn(svc)
	}

	return svc
}

func WithRoutingProcessorService(svc service.IRoutingProcessorService) BeneficiaryAccountServiceFunc {
	return func(s *BeneficiaryAccountService) {
		s.routingProcessorSvc = svc
	}
}

func WithConfig(cfg *config.Config) BeneficiaryAccountServiceFunc {
	return func(ds *BeneficiaryAccountService) {
		ds.config = cfg
	}
}
