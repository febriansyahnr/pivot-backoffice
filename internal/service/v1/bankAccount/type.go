package bankAccount

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

type bankAccountService struct {
	repo   repository.IBankAccountRepository
	logger logger.ILogger
}

type Option func(*bankAccountService)

var otelTracer = otel.Tracer("WithdrawalService")

func New(
	repo repository.IBankAccountRepository,
	logger logger.ILogger,
	options ...Option,
) service.IBankAccountService {
	svc := &bankAccountService{
		repo:   repo,
		logger: logger,
	}

	for _, opt := range options {
		opt(svc)
	}
	return svc
}
