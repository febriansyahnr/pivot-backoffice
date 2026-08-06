package settlementHold

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMSettlementHoldController")

type CRMSettlementHoldController struct {
	settlementHoldSvc service.ISettlementHoldService
	validate          *validator.Validate
}

func New(
	settlementHoldSvc service.ISettlementHoldService,
	validate *validator.Validate,
) controller.V1CRMSettlementController {
	c := &CRMSettlementHoldController{
		settlementHoldSvc: settlementHoldSvc,
		validate:          validate,
	}

	return c
}
