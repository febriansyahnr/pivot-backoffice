package internal_merchant

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalMerchantController")

type V1InternalMerchantController struct {
	forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService
	merchantSvc         service.IMerchantService
	validate            *validator.Validate
}

func New(forbiddenSvc service.IMerchantForbiddenUseCaseService,
	merchantSvc service.IMerchantService,
	validate *validator.Validate,
) controller.V1InternalMerchantController {
	return &V1InternalMerchantController{
		forbiddenUsecaseSvc: forbiddenSvc,
		merchantSvc:         merchantSvc,
		validate:            validate,
	}
}
