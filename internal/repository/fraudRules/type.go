package fraudrulesrepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("FraudRulesRepository")

const (
	tableName    = "fraud_rules"
	tableColumns = `fr.uuid, fr.rule_name, fr.condition, fr.priority, fr.weight, fr.is_active, fr.reference_type, fr.provider, fr.created_at, fr.updated_at, fr.deleted_at`
)

type FraudRulesRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	logger logger.ILogger,
	db mySqlExt.IMySqlExt,
) repository.IFraudRulesRepository {
	return &FraudRulesRepository{
		db:     db,
		logger: logger,
	}
}
