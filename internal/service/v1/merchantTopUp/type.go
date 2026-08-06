package merchantTopUp

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantTopUpService")

type merchantTopUpService struct {
	config            *config.Config
	logger            logger.ILogger
	paymentMethodRepo repository.IPaymentMethodRepository
	merchantTopUpRepo repository.IMerchantTopUpRepository
	snapCore          repository.ISnapCoreRepository
	merchantService   service.IMerchantService
	orchestratorSvc   service.IOrchestratorService
	rabbitMqExt       rabbitMqExt.IRabbitMQExt
	feeSvc            service.IFeeService

	internal service.IMerchantTopUpService
}

type OptionFunc func(*merchantTopUpService)

func New(
	config *config.Config,
	logger logger.ILogger,
	paymentMethodRepo repository.IPaymentMethodRepository,
	merchantTopUpRepo repository.IMerchantTopUpRepository,
	snapCore repository.ISnapCoreRepository,
	depends ...OptionFunc,
) service.IMerchantTopUpService {
	s := &merchantTopUpService{
		config: config,
		// secret:            secret,
		logger:            logger,
		paymentMethodRepo: paymentMethodRepo,
		merchantTopUpRepo: merchantTopUpRepo,
		snapCore:          snapCore,
	}
	s.internal = s

	for _, fn := range depends {
		fn(s)
	}
	return s
}

func WithOrchestratorService(svc service.IOrchestratorService) OptionFunc {
	return func(ds *merchantTopUpService) {
		ds.orchestratorSvc = svc
	}
}

func WithMerchantService(svc service.IMerchantService) OptionFunc {
	return func(ds *merchantTopUpService) {
		ds.merchantService = svc
	}
}

func WithRabbitMQClient(rmq rabbitMqExt.IRabbitMQExt) OptionFunc {
	return func(ds *merchantTopUpService) {
		ds.rabbitMqExt = rmq
	}
}

func WithInternalService(internal service.IMerchantTopUpService) OptionFunc {
	return func(ds *merchantTopUpService) {
		ds.internal = internal
	}
}

func WithFeeService(svc service.IFeeService) OptionFunc {
	return func(ds *merchantTopUpService) {
		ds.feeSvc = svc
	}
}
