package merchantForbiddenUsecase

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantForbiddenUsecaseService")

type MerchantForbiddenUseCaseService struct {
	logger          logger.ILogger
	repo            repository.IMerchantForbiddenUsecaseRepository
	rabbitMqExt     rabbitMqExt.IRabbitMQExt
	merchantService service.IMerchantService
}

func New(logger logger.ILogger,
	repo repository.IMerchantForbiddenUsecaseRepository,
	rmqExt rabbitMqExt.IRabbitMQExt,
	merchantService service.IMerchantService) *MerchantForbiddenUseCaseService {
	return &MerchantForbiddenUseCaseService{
		logger:          logger,
		repo:            repo,
		rabbitMqExt:     rmqExt,
		merchantService: merchantService,
	}
}
