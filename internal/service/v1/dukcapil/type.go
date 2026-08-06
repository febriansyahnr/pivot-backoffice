package dukcapilservice

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("DukcapilService")

type DukcapilServiceFunc func(*DukcapilService)

type DukcapilService struct {
	cfg                 *config.Config
	secret              *config.Secret
	logger              logger.ILogger
	dukcapilGatewayRepo repository.IDukcapilGatewayRepository
	merchantRepository  repository.IMerchantRepository
}

func New(
	cfg *config.Config,
	secret *config.Secret,
	log logger.ILogger,
	dukcapilGatewayRepo repository.IDukcapilGatewayRepository,
	merchantRepository repository.IMerchantRepository,
	depends ...DukcapilServiceFunc,
) service.IDukcapilService {
	service := &DukcapilService{
		cfg:                 cfg,
		secret:              secret,
		logger:              log,
		dukcapilGatewayRepo: dukcapilGatewayRepo,
		merchantRepository:  merchantRepository,
	}

	for _, fn := range depends {
		fn(service)
	}

	return service
}

func (s *DukcapilService) GetFieldThresholds() config.DukcapilFieldThresholds {
	dukcapilFF := constant.GetDukcapilFeatureFlag(s.cfg.Environment, s.logger)
	if dukcapilFF == nil {
		return s.cfg.Dukcapil.FieldThresholds
	}

	return dukcapilFF.FieldThresholds
}
