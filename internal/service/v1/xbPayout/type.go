package xbPayoutService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("XbPayoutService")

const serviceName = "XbPayout-Service"

type xbPayoutService struct {
	logger                 logger.ILogger
	disbursementRepo       repository.IDisbursementRepository
	beneficiaryAccountRepo repository.IBeneficiaryAccountRepository
	xbCoreProcessorRepo    repository.IXbCoreProcessorRepository
	statusHistoriesRepo    repository.IStatusHistoriesRepository

	feeSvc          service.IFeeService
	orchestratorSvc service.IOrchestratorService

	rabbitMqExt rabbitMqExt.IRabbitMQExt
	gcs         gcs.IGCSService
	config      *config.Config
}

type XbPayoutServiceFunc func(*xbPayoutService)

func New(
	logger logger.ILogger,
	disbursementRepo repository.IDisbursementRepository,
	beneficiaryAccountRepo repository.IBeneficiaryAccountRepository,
	xbCoreProcessorRepo repository.IXbCoreProcessorRepository,
	depends ...XbPayoutServiceFunc,
) service.IXbPayoutService {
	s := &xbPayoutService{
		logger:                 logger,
		disbursementRepo:       disbursementRepo,
		beneficiaryAccountRepo: beneficiaryAccountRepo,
		xbCoreProcessorRepo:    xbCoreProcessorRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithFeeService(svc service.IFeeService) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.feeSvc = svc
	}
}

func WithOrchestratorService(svc service.IOrchestratorService) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.orchestratorSvc = svc
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.rabbitMqExt = rmq
	}
}

func WithConfig(c *config.Config) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.config = c
	}
}

func WithGCS(gcsSvc gcs.IGCSService) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.gcs = gcsSvc
	}
}

func WithStatusHistories(statusHistories repository.IStatusHistoriesRepository) XbPayoutServiceFunc {
	return func(ds *xbPayoutService) {
		ds.statusHistoriesRepo = statusHistories
	}
}
