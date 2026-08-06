package transfer

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("TransferInternalController")

type TransferInternalController struct {
	transferService service.ITransferService
	validator       *validator.Validate
}

func New(
	transferSvc service.ITransferService,
	validator *validator.Validate,
) controller.V1InternalTransferController {
	return &TransferInternalController{
		transferService: transferSvc,
		validator:       validator,
	}
}
