package crmCardFundedPayoutController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMCardFundedPayoutController")

type handler struct {
	validate *validatorExt.Validate
	service  service.ICardFundedPayoutService
}

func New(vld *validatorExt.Validate, svc service.ICardFundedPayoutService) controller.V1CRMCardFundedPayoutController {
	return &handler{vld, svc}
}
