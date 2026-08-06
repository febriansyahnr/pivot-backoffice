package payoutManualProcessingAccount

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("PayoutManualProcessingAccountService")

type PayoutManualProcessingAccountService struct {
	repo   repository.IPayoutManualProcessingAccountRepository
	logger logger.ILogger
}

func New(
	repo repository.IPayoutManualProcessingAccountRepository,
	logger logger.ILogger,
) service.IPayoutManualProcessingAccountService {
	return &PayoutManualProcessingAccountService{
		repo:   repo,
		logger: logger,
	}
}
