package processorCallbackController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ProcessorCallbackController")

type processorCallbackController struct {
	logger              logger.ILogger
	disbursementSvc     service.IDisbursementService
	routingProcessorSvc service.IRoutingProcessorService
	accountInquirySvc   service.IAccountInquiryService
	validator           *validatorExt.Validate
}

func New(
	logger logger.ILogger,
	disbursementSvc service.IDisbursementService,
	routingProcessorSvc service.IRoutingProcessorService,
	accountInquirySvc service.IAccountInquiryService,
	validator *validatorExt.Validate,
) controller.V1ProcessorCallbackController {
	return &processorCallbackController{
		logger:              logger,
		disbursementSvc:     disbursementSvc,
		validator:           validator,
		accountInquirySvc:   accountInquirySvc,
		routingProcessorSvc: routingProcessorSvc,
	}
}
