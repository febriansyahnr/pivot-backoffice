package routingprocessorService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("RoutingProcessorService")

type RoutingProcessorServiceFunc func(*routingProcessorService)

type routingProcessorService struct {
	cfg                       *config.Config
	logger                    logger.ILogger
	routingProcessor          map[string]repository.IRoutingProcessorRepository // mapping for common functionalities at third party
	requestAccountInquiryRepo repository.IRequestAccountInquiryRepository
	outboundRepository        repository.IOutboundRepository
	rabbitMq                  rabbitMqExt.IRabbitMQExt
	flipProcessorRepository   repository.IFlipProcessorRepository
	danaProcessorRepository   repository.IDanaProcessorRepository
}

func New(
	cfg *config.Config,
	log logger.ILogger,
	routingProcessor map[string]repository.IRoutingProcessorRepository,
	depends ...RoutingProcessorServiceFunc,
) service.IRoutingProcessorService {
	service := &routingProcessorService{
		cfg:              cfg,
		logger:           log,
		routingProcessor: routingProcessor,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}

func WithRabbitMqExt(rabbitMq rabbitMqExt.IRabbitMQExt) RoutingProcessorServiceFunc {
	return func(rs *routingProcessorService) {
		rs.rabbitMq = rabbitMq
	}
}

func WithOutboundRepository(outboundRepository repository.IOutboundRepository) RoutingProcessorServiceFunc {
	return func(rs *routingProcessorService) {
		rs.outboundRepository = outboundRepository
	}
}

func WithRequestAccountInquiryRepository(repo repository.IRequestAccountInquiryRepository) RoutingProcessorServiceFunc {
	return func(rps *routingProcessorService) {
		rps.requestAccountInquiryRepo = repo
	}
}

func WithFlipProcessorRepository(repo repository.IFlipProcessorRepository) RoutingProcessorServiceFunc {
	return func(rps *routingProcessorService) {
		rps.flipProcessorRepository = repo
	}
}

func WithDanaProcessorRepository(repo repository.IDanaProcessorRepository) RoutingProcessorServiceFunc {
	return func(rps *routingProcessorService) {
		rps.danaProcessorRepository = repo
	}
}
