package payout

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PayoutCron")

type PayoutHandlerOption func(*cronHandler)

type cronHandler struct {
	config  *config.Config
	log     logger.ILogger
	service service.IDisbursementService
}

func New(log logger.ILogger, service service.IDisbursementService, opts ...PayoutHandlerOption) handler.IPayoutHandler {
	handler := &cronHandler{log: log, service: service}

	for _, opt := range opts {
		opt(handler)
	}

	return handler
}

func WithConfig(config *config.Config) PayoutHandlerOption {
	return func(h *cronHandler) {
		h.config = config
	}
}
