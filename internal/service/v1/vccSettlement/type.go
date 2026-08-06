package vccsettlement

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("VccSettlementService")

type VccSettlementService struct {
	config            config.Config
	logger            logger.ILogger
	merchantRcnSvc    service.IMerchantRcnService
	cimbProcessor     repository.ICimbProcessorRepository
	vccSettlementRepo repository.IVCCSettlementRepository
	notificationSvc   service.INotificationService
	cache             redisExt.IRedisExt
	rabbitMq          rabbitMqExt.IRabbitMQExt
}

func New(
	config config.Config,
	logger logger.ILogger,
	merchantRcnSvc service.IMerchantRcnService,
	notificationSvc service.INotificationService,
	cimbProcessor repository.ICimbProcessorRepository,
	vccSettlementRepo repository.IVCCSettlementRepository,
	cache redisExt.IRedisExt,
	rabbitMq rabbitMqExt.IRabbitMQExt,
) service.IVccSettlementService {
	return &VccSettlementService{
		config:            config,
		logger:            logger,
		merchantRcnSvc:    merchantRcnSvc,
		notificationSvc:   notificationSvc,
		cimbProcessor:     cimbProcessor,
		vccSettlementRepo: vccSettlementRepo,
		cache:             cache,
		rabbitMq:          rabbitMq,
	}
}
