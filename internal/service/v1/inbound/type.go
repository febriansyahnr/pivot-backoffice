package inboundService

import (
	"go.opentelemetry.io/otel"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	pdkEncoder "github.com/paper-indonesia/pdk/v2/logger/encoder"
)

var otelTracer = otel.Tracer("InboundService")

type InboundService struct {
	cfg         *config.Config
	logger      pdkLogger.ILogger
	inboundRepo repository.IInboundRepository
	inspector   pdkEncoder.Inspector
}

type InboundServiceFunc func(*InboundService)

func New(cfg *config.Config, logger pdkLogger.ILogger, inboundRepo repository.IInboundRepository, depends ...InboundServiceFunc) services.IInboundService {
	s := &InboundService{
		cfg:         cfg,
		logger:      logger,
		inboundRepo: inboundRepo,
	}

	for _, fn := range depends {
		fn(s)
	}

	return s
}

func WithMaskingSensitiveData(fields []string) InboundServiceFunc {
	return func(s *InboundService) {
		s.inspector = pdkEncoder.NewInspector(fields)
	}
}
