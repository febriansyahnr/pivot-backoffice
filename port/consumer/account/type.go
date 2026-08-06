package account

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AccountConsumer")

type AccountConsumer struct {
	logger     logger.ILogger
	AccountSvc service.IAccountService
}

func New(
	logger logger.ILogger,
	accountSvc service.IAccountService) *AccountConsumer {
	return &AccountConsumer{
		logger:     logger,
		AccountSvc: accountSvc,
	}
}
