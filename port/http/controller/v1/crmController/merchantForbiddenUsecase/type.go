package merchantForbiddenUsecase

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1CRMMerchantForbiddenUseCaseController")

type V1CRMMerchantForbiddenUseCaseController struct {
	forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService
	validate            *validator.Validate
}

func New(forbiddenSvc service.IMerchantForbiddenUseCaseService,
	validate *validator.Validate,
) *V1CRMMerchantForbiddenUseCaseController {
	return &V1CRMMerchantForbiddenUseCaseController{
		forbiddenUsecaseSvc: forbiddenSvc,
		validate:            validate,
	}
}
