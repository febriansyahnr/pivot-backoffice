package amlservice

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("AmlService")

type AmlServiceFunc func(*AmlService)

type AmlService struct {
	cfg                 *config.Config
	logger              logger.ILogger
	merchantRepository  repository.IMerchantRepository
	thirdPartyProcessor map[string]repository.IAmlProcessorRepository // mapping for common functionalities at third party
	outboundRepository  repository.IOutboundRepository
	rabbitMq            rabbitMqExt.IRabbitMQExt
}

func New(
	cfg *config.Config,
	log logger.ILogger,
	merchantRepository repository.IMerchantRepository,
	thirdPartyProcessor map[string]repository.IAmlProcessorRepository,
	depends ...AmlServiceFunc,
) service.IAmlService {
	service := &AmlService{
		cfg:                 cfg,
		logger:              log,
		merchantRepository:  merchantRepository,
		thirdPartyProcessor: thirdPartyProcessor,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}

func WithRabbitMqExt(rabbitMq rabbitMqExt.IRabbitMQExt) AmlServiceFunc {
	return func(rs *AmlService) {
		rs.rabbitMq = rabbitMq
	}
}

func WithOutboundRepository(outboundRepository repository.IOutboundRepository) AmlServiceFunc {
	return func(rs *AmlService) {
		rs.outboundRepository = outboundRepository
	}
}
