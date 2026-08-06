package depositSettingController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("DepositSettingController")

type handler struct {
	validator   *validatorExt.Validate
	logger      logger.ILogger
	merchantSvc service.IMerchantService
}

func New(vld *validatorExt.Validate, log logger.ILogger, merchantSvc service.IMerchantService) controller.V1DepositSettingController {
	return &handler{
		validator: vld, logger: log, merchantSvc: merchantSvc,
	}
}
