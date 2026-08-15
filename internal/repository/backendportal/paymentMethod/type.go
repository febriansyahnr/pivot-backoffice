package paymentMethodRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

const (
	tableName                  = "payment_methods"
	tablePaymentMethodMerchant = "payment_method_merchant"
)

var otelTracer = otel.Tracer("PaymentMethodRepository")

type PaymentMethodRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IPaymentMethodRepository {
	return &PaymentMethodRepository{
		db:     db,
		logger: logger,
	}
}

const (
	SelectStrPaymentMethod = `uuid, type, sub_type, category, name, description, logo, acquirer, bank_name, instructions,
					processor, activation_method, country_of_operation, supported_currency, config, required_document,
       				created_at, updated_at, deleted_at`

	SelectPaymentMethodWithPivotStr = "p.uuid, p.type, p.sub_type, p.category, p.name, p.description, p.logo, p.acquirer, p.bank_name, p.instructions, p.processor, p.activation_method, p.country_of_operation, p.supported_currency, p.config, p.created_at, p.updated_at, p.deleted_at, COALESCE(pmm.merchant_id, ?) as merchant_id, COALESCE(pmm.is_active, false) as is_active, CASE WHEN pmm.activation_status IS NOT NULL THEN pmm.activation_status WHEN p.activation_method = 'INSTANT' THEN 'APPROVED' ELSE 'NOT_REQUESTED' END as activation_status, COALESCE(pmm.channel_type, 'AGGREGATOR') as channel_type, pmm.config as merchant_config, p.required_document"
)
