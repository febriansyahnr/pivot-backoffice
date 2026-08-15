package disbursementRepository

import (
	"time"

	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

const (
	SelectDisbursementWithTransactionStr = `d.uuid, d.reference_id, d.merchant_id, d.bulk_id, d.purpose_id, d.sender_name, d.account_inquiry_id, 
		d.beneficiary_bank_code, d.beneficiary_bank_name, d.beneficiary_account_no, d.beneficiary_account_name, d.processor_reference_id, 
		d.bank_reference_no, d.currency, d.amount, d.fee, d.total_amount, d.status, d.reason_type, d.reason_description, d.remark, d.created_from, 
		COALESCE(c.name, 'System') as created_by, COALESCE(a.name, '-') as approved_by, d.approved_at, d.created_at, d.updated_at,
		IF(IFNULL(d.reason_type, '') = 'REVERSAL', d.reason_type, t.status) AS transaction_status, t.reason_type as transaction_reason_type, t.reason_description as transaction_reason_description, d.metadata, t.processor_reference, t.processor_reference_id as transaction_processor_reference_id`
)

var (
	otelTracer = otel.Tracer("DisbursementRepository")
	local, _   = time.LoadLocation(constant.TimeLoc)
)

const tableName = "disbursements"

type Option func(*DisbursementRepository)

type DisbursementRepository struct {
	pdkLogger pdkLogger.ILogger
	db        mySqlExt.IMySqlExt
	config    *config.DisbursementConfig
	appConfig *config.AppConfig
}

func New(
	db mySqlExt.IMySqlExt,
	pdkLogger pdkLogger.ILogger,
	options ...Option,
) repository.IDisbursementRepository {
	r := &DisbursementRepository{
		pdkLogger: pdkLogger,
		db:        db,
	}
	for _, opt := range options {
		opt(r)
	}
	return r
}

func WithConfig(config *config.DisbursementConfig) Option {
	return func(dr *DisbursementRepository) {
		dr.config = config
	}
}

func WithAppConfig(config *config.AppConfig) Option {
	return func(dr *DisbursementRepository) {
		dr.appConfig = config
	}
}
