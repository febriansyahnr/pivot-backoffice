package merchantRcn

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("MerchantRcnController")

type MerchantRcnController struct {
	merchantSvc      service.IMerchantRcnService
	vccSettlementSvc service.IVccSettlementService
	validate         *validator.Validate
}

type ControllerFunc func(*MerchantRcnController)

func New(
	merchantSvc service.IMerchantRcnService,
	validate *validator.Validate,
	depends ...ControllerFunc) *MerchantRcnController {
	c := &MerchantRcnController{
		merchantSvc: merchantSvc,
		validate:    validate,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithVccSettlementService(svc service.IVccSettlementService) ControllerFunc {
	return func(mrc *MerchantRcnController) {
		mrc.vccSettlementSvc = svc
	}
}
