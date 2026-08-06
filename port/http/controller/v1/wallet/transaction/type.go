package walletTransaction

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	"github.com/paper-indonesia/pivot-backoffice/port/http/controller"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("WalletTransactionController")

const (
	maxRangeDays      = 31
	maxBackdateDays   = 182
	maxBackdateMonths = 6 // This value is the month format of the constant maxBackdateDays
)

type handler struct {
	vld     *validatorExt.Validate
	service service.IWalletTransactionService
}

func New(vld *validatorExt.Validate, svc service.IWalletTransactionService) controller.V1WalletTransactionController {
	return &handler{vld, svc}
}
