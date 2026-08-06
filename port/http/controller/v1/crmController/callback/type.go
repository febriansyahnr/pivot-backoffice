package crmCallbackController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMCallbackController")

type handler struct {
	validator         *validatorExt.Validate
	logger            logger.ILogger
	unifiedPaymentSvc service.IUnifiedPaymentService
	disbursementSvc   service.IDisbursementService
}

func New(
	logger logger.ILogger,
	unifiedPaymentSvc service.IUnifiedPaymentService,
	disbursementSvc service.IDisbursementService,
) controller.V1CRMCallbackController {
	return &handler{
		validator:         validatorExt.New(),
		logger:            logger,
		unifiedPaymentSvc: unifiedPaymentSvc,
		disbursementSvc:   disbursementSvc,
	}
}
