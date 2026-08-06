package tnc

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("TNCRepository")

const (
	versionsTableName    = "tncs"
	versionsTableColumns = `t.uuid, t.version, t.title, t.markdown_content, t.is_active, t.created_by, t.created_at, t.updated_at, t.deleted_at`

	historyTableName    = "merchant_tnc_signing_histories"
	historyTableColumns = `t.uuid, t.merchant_id, t.tnc_id, t.version, t.signed_by, t.signed_by_email, t.signed_at, t.document_path, t.ip_address, t.user_agent, t.created_at`
)

type TNCRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(logger logger.ILogger, db mySqlExt.IMySqlExt) repository.ITNCRepository {
	return &TNCRepository{
		Properties: basicsql.NewBasicSQLProperties(db),
		db:         db,
		logger:     logger,
	}
}
