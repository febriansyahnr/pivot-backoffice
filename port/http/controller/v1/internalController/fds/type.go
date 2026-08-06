package fds

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InternalFdsController")

type InternalFdsController struct {
	config   *config.Config
	logger   logger.ILogger
	validate *validator.Validate
	fdsSvc   service.IFdsService
}

type OptionFunc func(*InternalFdsController)

func New(
	config *config.Config,
	logger logger.ILogger,
	validate *validator.Validate,
	fdsSvc service.IFdsService,
	opts ...OptionFunc,
) controller.V1InternalFdsController {
	handler := &InternalFdsController{
		config:   config,
		logger:   logger,
		validate: validate,
		fdsSvc:   fdsSvc,
	}

	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

func (c *InternalFdsController) GetTimeout() int64 {
	fdsFF := constant.GetFdsFeatureFlag(c.config.Environment)
	if fdsFF == nil {
		return c.config.FdsConfig.Timeout
	}

	return fdsFF.Timeout
}
