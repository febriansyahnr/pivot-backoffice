package crmDisbursementController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMDisbursementController")

type handler struct {
	validator           *validatorExt.Validate
	disbursementSvc     service.IDisbursementService
	routingProcessorSvc service.IRoutingProcessorService
}

func New(disbursementSvc service.IDisbursementService, routingProcessorSvc service.IRoutingProcessorService) controller.V1CRMDisbursementController {
	return &handler{
		disbursementSvc:     disbursementSvc,
		routingProcessorSvc: routingProcessorSvc,
		validator:           validatorExt.New(),
	}
}
