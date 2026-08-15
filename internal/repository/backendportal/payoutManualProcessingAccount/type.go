package payoutManualProcessingAccount

import (
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("PayoutManualProcessingAccountRepository")

const (
	tableName    = "payout_manual_processing_accounts"
	tableColumns = `a.uuid, a.merchant_id, a.bank_code, a.account_number, a.status`
	// listColumns joins merchants to resolve merchant_name; only used by List.
	listColumns = `a.uuid, m.name AS merchant_name, a.merchant_id, a.bank_code, a.account_number, a.status, a.updated_by, a.updated_at`
)

type PayoutManualProcessingAccountRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	logger logger.ILogger,
	db mySqlExt.IMySqlExt,
) repository.IPayoutManualProcessingAccountRepository {
	return &PayoutManualProcessingAccountRepository{
		db:     db,
		logger: logger,
	}
}
