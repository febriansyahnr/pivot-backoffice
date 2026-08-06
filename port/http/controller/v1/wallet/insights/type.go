package walletInsightsController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("WalletInsightsController")
)

type WalletInsightsController struct {
	insightSvc service.IWalletInsightService
}

func New(
	insightSvc service.IWalletInsightService,
) controller.V1WalletInsightController {
	return &WalletInsightsController{
		insightSvc: insightSvc,
	}
}
