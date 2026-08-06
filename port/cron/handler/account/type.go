package account

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/port/cron/handler"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("AccountCronHandler")

type account struct {
	logger     logger.ILogger
	accountSvc service.IAccountService
}

func NewAccount(
	logger logger.ILogger,
	accountSvc service.IAccountService,
) handler.IOrchestratorBalanceHandler {
	return &account{
		logger:     logger,
		accountSvc: accountSvc,
	}
}
