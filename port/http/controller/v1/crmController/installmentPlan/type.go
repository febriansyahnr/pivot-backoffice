package installmentplan

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1InstallmentPlanController")

type V1CRMInstallmentPlanController struct {
	validate           *validator.Validate
	installmentPlanSvc service.IInstallmentPlanService
}

func NewController(
	installmentPlanSvc service.IInstallmentPlanService,
	validate *validator.Validate,
) controller.V1CRMInstallmentPlanController {
	return &V1CRMInstallmentPlanController{
		validate:           validate,
		installmentPlanSvc: installmentPlanSvc,
	}
}
