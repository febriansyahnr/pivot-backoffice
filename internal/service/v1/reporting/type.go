package reportingService

import (
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	port "github.com/paper-indonesia/pivot-backoffice/internal/service"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("ReportingService")
	tz, _      = time.LoadLocation(constant.TimeLoc)
)

type service struct {
	logger      logger.ILogger
	repo        repository.IReportingRepository
	accountRepo repository.IAccountRepository
}

func New(log logger.ILogger, repo repository.IReportingRepository, accountRepo repository.IAccountRepository) port.IReportingService {
	return &service{logger: log, repo: repo, accountRepo: accountRepo}
}
