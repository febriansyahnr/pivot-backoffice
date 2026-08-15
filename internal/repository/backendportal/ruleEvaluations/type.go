package ruleevaluationsrepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("RuleEvaluationsRepository")

const tableName = "rule_evaluations"

type RuleEvaluationsRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	logger logger.ILogger,
	db mySqlExt.IMySqlExt,
) repository.IRuleEvaluationsRepository {
	return &RuleEvaluationsRepository{
		db:     db,
		logger: logger,
	}
}
