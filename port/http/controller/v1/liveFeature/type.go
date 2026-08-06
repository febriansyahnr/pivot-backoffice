package liveFeature

import (
	"go.opentelemetry.io/otel"

	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
)

var otelTracer = otel.Tracer("LiveFeatureController")

type LiveFeatureController struct {
	featureSvc service.ILiveFeaturesService
}

func New(featureSvc service.ILiveFeaturesService) controller.V1LiveFeatureController {
	return &LiveFeatureController{
		featureSvc: featureSvc,
	}
}
