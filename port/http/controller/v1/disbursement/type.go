package disbursementController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/gcs"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/go/monitoring"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DisbursementController")

type Controller struct {
	config                   *config.Config
	validate                 *validator.Validate
	monitor                  *monitoring.Monitor
	merchant                 service.IMerchantService
	disbursementDashboardSvc service.IDisbursementDashboardService
	disbursementSvc          service.IDisbursementService
	beneficiaryAccountSvc    service.IBeneficiaryAccountService
	feeSvc                   service.IFeeService

	rabbitMqExt rabbitMqExt.IRabbitMQExt
	gcs         gcs.IGCSService
	redis       redisExt.IRedisExt
	logger      logger.ILogger
}

type ControllerFunc func(*Controller)

type Services struct {
	MerchantSvc              service.IMerchantService
	DisbursementDashboardSvc service.IDisbursementDashboardService
	DisbursementSvc          service.IDisbursementService
	BeneficiaryAccountSvc    service.IBeneficiaryAccountService
	FeeSvc                   service.IFeeService
}

func New(
	config *config.Config,
	validate *validator.Validate,
	monitor *monitoring.Monitor,
	services Services,
	rabbitMqExt rabbitMqExt.IRabbitMQExt,
	gcs gcs.IGCSService,
	depends ...ControllerFunc,
) controller.V1DisbursementController {
	c := &Controller{
		config:                   config,
		validate:                 validate,
		monitor:                  monitor,
		merchant:                 services.MerchantSvc,
		disbursementDashboardSvc: services.DisbursementDashboardSvc,
		disbursementSvc:          services.DisbursementSvc,
		beneficiaryAccountSvc:    services.BeneficiaryAccountSvc,
		feeSvc:                   services.FeeSvc,
		rabbitMqExt:              rabbitMqExt,
		gcs:                      gcs,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithRedisClient(rdb redisExt.IRedisExt) ControllerFunc {
	return func(c *Controller) {
		c.redis = rdb
	}
}

func WithLogger(log logger.ILogger) ControllerFunc {
	return func(c *Controller) {
		c.logger = log
	}
}

var (
	minAmountErrFmt = "Min amount is Rp %s"
	maxAmountErrFmt = "Maximum amount is Rp %s"
)
