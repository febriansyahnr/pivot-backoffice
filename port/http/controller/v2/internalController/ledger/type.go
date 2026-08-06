package ledgerController

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("LedgerController")

type LedgerController struct {
	ledgerSvc service.ILedgerService
	validator *validatorExt.Validate
}

func New(ledgersvc service.ILedgerService) controller.V2InternalLedgerController {
	return &LedgerController{
		ledgerSvc: ledgersvc,
		validator: validatorExt.New(),
	}
}
