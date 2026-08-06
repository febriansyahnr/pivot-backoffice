package withdrawalCrmController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"go.opentelemetry.io/otel"
)

type handler struct {
	validator *validatorExt.Validate
	service   service.IWithdrawalService
}

var otelTracer = otel.Tracer("WithdrawalCRMController")

func New(vld *validatorExt.Validate, svc service.IWithdrawalService) controller.V1CRMWithdrawalController {
	return &handler{validator: vld, service: svc}
}
