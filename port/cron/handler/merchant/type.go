package merchantCronHandler

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"
	"go.opentelemetry.io/otel"
	"time"
)

var (
	otelTracer = otel.Tracer("MerchantCronHandler")
	tz, _      = time.LoadLocation(constant.TimeLoc)
)

type merchantCronHandler struct {
	logger      logger.ILogger
	merchantSvc service.IMerchantService
}

func New(
	logger logger.ILogger,
	merchantSvc service.IMerchantService,
) handler.IMerchantHandler {
	return &merchantCronHandler{
		logger, merchantSvc}
}
