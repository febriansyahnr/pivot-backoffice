package internalPayoutController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalPayoutController")

type InternalPayoutController struct {
	config                *config.Config
	validate              *validator.Validate
	disbursementSvc       service.IDisbursementService
	merchantSvc           service.IMerchantService
	inquiryAccSvc         service.IAccountInquiryService
	beneficiaryAccountSvc service.IBeneficiaryAccountService
	rabbitMqExt           rabbitMqExt.IRabbitMQExt
	redis                 redisExt.IRedisExt
	logger                logger.ILogger
}

type ControllerFunc func(*InternalPayoutController)

func New(
	config *config.Config,
	validate *validator.Validate,
	disbursementSvc service.IDisbursementService,
	merchantSvc service.IMerchantService,
	inquiryAccSvc service.IAccountInquiryService,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	depends ...ControllerFunc,
) controller.V1InternalPayoutController {
	c := &InternalPayoutController{
		validate:        validate,
		disbursementSvc: disbursementSvc,
		merchantSvc:     merchantSvc,
		rabbitMqExt:     rabbitMqExt,
		config:          config,
		inquiryAccSvc:   inquiryAccSvc,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithRedisClient(rdb redisExt.IRedisExt) ControllerFunc {
	return func(c *InternalPayoutController) {
		c.redis = rdb
	}
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *InternalPayoutController) {
		c.logger = log
	}
}

func WithBeneficiaryAccountService(svc service.IBeneficiaryAccountService) ControllerFunc {
	return func(c *InternalPayoutController) {
		c.beneficiaryAccountSvc = svc
	}
}
