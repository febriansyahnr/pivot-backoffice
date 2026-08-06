package accounttransaction_repository

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/tracer"

	"github.com/paper-indonesia/pdk/v2/logger"
)

var otelTracer = tracer.New("AccountTransactionRepository")

type Option func(*AccountTransactionRepository)
type AccountTransactionRepository struct {
	*basicsql.Properties

	db        mySqlExt.IMySqlExt
	logger    logger.ILogger
	appConfig *config.AppConfig
}

func New(
	db mySqlExt.IMySqlExt,
	logger logger.ILogger,
	options ...Option,
) repository.IAccountTransactionRepository {
	r := &AccountTransactionRepository{
		Properties: basicsql.NewBasicSQLProperties(db),

		db:     db,
		logger: logger,
	}

	for _, opt := range options {
		opt(r)
	}

	return r
}

func WithAppConfig(config *config.AppConfig) Option {
	return func(dr *AccountTransactionRepository) {
		dr.appConfig = config
	}
}

const tableName = "account_transactions"

const queryWithUseCase = `SELECT 
			t.uuid, 
			t.merchant_id, 
			t.reference_id, 
			t.account_id,       
			t.currency, 
			t.credit, 
			t.debit, 
			CASE 
				WHEN t.type = 'DISBURSEMENT' AND d.bulk_id IS NOT NULL THEN 'BULK_DISBURSEMENT'  
				ELSE t.type
			END AS type, 
			t.status, 
			t.channel,
			t.reason_type, 
			t.reason_description, 
			t.remarks, 
			t.transaction_timestamp,
			t.settlement_at,
			t.settlement_status,
			t.settlement_model,
			t.created_at, 
			t.updated_at,
			COALESCE(d.sender_name, '-') as sender_name,
			CASE 
				WHEN d.fee IS NOT NULL THEN d.fee 
				WHEN p.fee IS NOT NULL THEN p.fee 
				ELSE 0 
			END as fee,
			COALESCE(d.beneficiary_account_no, '-') as beneficiary_account_no,
			COALESCE(d.beneficiary_account_name, '-') as beneficiary_account_name,
			COALESCE(d.beneficiary_bank_name, '-') as beneficiary_bank_name,
			CASE 
				WHEN d.reference_id IS NOT NULL THEN d.reference_id 
				WHEN p.reference_id IS NOT NULL THEN p.reference_id 
				ELSE '-' 
			END as client_reference_id,
			d.approved_at,
			COALESCE(t.processor_reference_id, '-') as processor_reference_id,
			COALESCE(c.name, 'System') as created_by,
			COALESCE(a.name, '-') as approved_by, t.additional_info, t.reference
        FROM account_transactions t
        LEFT JOIN disbursements d ON d.uuid = t.reference_id AND t.type = 'DISBURSEMENT'
        LEFT JOIN payments p ON p.uuid = t.reference_id AND t.type = 'PAYMENT'
        LEFT JOIN users c ON c.uuid = d.created_by
        LEFT JOIN users a ON a.uuid = d.approved_by`
