package crmfdscontroller

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("CRMFraudRuleController")

type CRMFdsController struct {
	config     *config.Config
	logger     logger.ILogger
	validate   *validator.Validate
	fdsService service.IFdsService
}

type OptionFunc func(*CRMFdsController)

func New(
	config *config.Config,
	logger logger.ILogger,
	validate *validator.Validate,
	fdsService service.IFdsService,
	opts ...OptionFunc,
) controller.V1CRMFdsController {
	handler := &CRMFdsController{
		config:     config,
		logger:     logger,
		validate:   validate,
		fdsService: fdsService,
	}

	for _, opt := range opts {
		opt(handler)
	}
	return handler
}
