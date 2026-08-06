package installmentPlan

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("InstallmentPlanRepository")

const (
	installmentPlansTableName = "installment_plans"
)

type InstallmentPlanRepository struct {
	logger logger.ILogger
	db     mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IInstallmentPlanRepository {
	return &InstallmentPlanRepository{db: db, logger: logger}
}
