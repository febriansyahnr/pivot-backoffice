package vendor

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("VendorRepository")

const (
	tableName    = "vendors"
	tableColumns = `v.uuid, v.merchant_id, v.name, v.beneficial_owner, v.business_category, v.avg_monthly_tpv_amount, v.bank_name, v.bank_code, v.account_number, v.account_name, v.documents, v.status, v.created_at, v.updated_at, v.deleted_at`
)

type VendorRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(
	logger logger.ILogger,
	db mySqlExt.IMySqlExt,
) repository.IVendorRepository {
	return &VendorRepository{
		db:     db,
		logger: logger,
	}
}
