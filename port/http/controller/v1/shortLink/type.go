package shortlink

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("ShortLinkController")

type ShortLinkController struct {
	config       *config.Config
	shortLinkSvc service.IShortLinkService
}

func New(
	config *config.Config,
	shortLinkSvc service.IShortLinkService,
) *ShortLinkController {
	return &ShortLinkController{
		config:       config,
		shortLinkSvc: shortLinkSvc,
	}
}
