package apiLog

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("V1ApiLogsSettingController")

type handler struct {
	inboundSvc service.IInboundService
}

func New(inboundSvc service.IInboundService) controller.V1ApiLogsSettingController {
	return &handler{
		inboundSvc: inboundSvc,
	}
}
