package crmfraudrulecontroller

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMFraudRuleController")

type CRMFraudRuleController struct {
	validate         *validator.Validate
	fraudRuleService service.IFraudRuleService
}

func New(
	fraudRuleService service.IFraudRuleService,
	validate *validator.Validate,
) controller.V1CRMFraudRuleController {
	return &CRMFraudRuleController{
		validate:         validate,
		fraudRuleService: fraudRuleService,
	}
}
