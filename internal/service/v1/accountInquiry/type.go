package accountinquiry

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/snap/bankTransfer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	bankDB     = bankTransfer.NewBankDB()
	otelTracer = otel.Tracer("AccountInquiryService")
)

type AccountInquiryServiceFunc func(*AccountInquiryService)

type AccountInquiryService struct {
	logger   logger.ILogger
	rabbitMq rabbitMqExt.IRabbitMQExt

	snapCore            repository.ISnapCoreRepository
	repo                repository.IRequestAccountInquiryRepository
	accountInquiryRepo  repository.IAccountInquiriesRepository
	merchantService     service.IMerchantService
	orchestratorService service.IOrchestratorService
	feeSvc              service.IFeeService
	routingProcessorSvc service.IRoutingProcessorService
	beneficiaryRepo     repository.IBeneficiaryAccountRepository
	transferSvc         service.ITransferService
	outboundRepository  repository.IOutboundRepository

	config *config.Config
}

func New(
	logger logger.ILogger,
	snapCore repository.ISnapCoreRepository,
	repo repository.IRequestAccountInquiryRepository,
	accountInquiryRepo repository.IAccountInquiriesRepository,
	orchestratorService service.IOrchestratorService,
	merchantService service.IMerchantService,
	feeSvc service.IFeeService,
	depends ...AccountInquiryServiceFunc,
) service.IAccountInquiryService {
	service := &AccountInquiryService{
		logger:              logger,
		snapCore:            snapCore,
		orchestratorService: orchestratorService,
		repo:                repo,
		accountInquiryRepo:  accountInquiryRepo,
		merchantService:     merchantService,
		feeSvc:              feeSvc,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}

func WithBeneficiaryAccountRepository(repo repository.IBeneficiaryAccountRepository) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.beneficiaryRepo = repo
	}
}

func WithRoutingProcessorService(service service.IRoutingProcessorService) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.routingProcessorSvc = service
	}
}

func WithConfig(cfg *config.Config) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.config = cfg
	}
}

func WithTransferService(svc service.ITransferService) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.transferSvc = svc
	}
}

func WithOutboundRepository(repo repository.IOutboundRepository) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.outboundRepository = repo
	}
}

func WithRabbitMqExt(rmq rabbitMqExt.IRabbitMQExt) AccountInquiryServiceFunc {
	return func(ds *AccountInquiryService) {
		ds.rabbitMq = rmq
	}
}
