package internalWithdrawalsController

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("internalWithdrawalController")

type InternalWithdrawalController struct {
	config   *config.Config
	validate *validator.Validate

	withdrawalSvc service.IWithdrawalService
}

type ControllerFunc func(*InternalWithdrawalController)

func New(
	config *config.Config,
	validate *validator.Validate,
	depends ...ControllerFunc,
) controller.V1InternalWithdrawalController {
	c := &InternalWithdrawalController{
		config:   config,
		validate: validate,
	}
	for _, fn := range depends {
		fn(c)
	}
	return c
}

func WithWithdrawalService(svc service.IWithdrawalService) ControllerFunc {
	return func(c *InternalWithdrawalController) {
		c.withdrawalSvc = svc
	}
}
