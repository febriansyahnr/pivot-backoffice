package accountInquiries

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

const tableName = "account_inquiries"

var otelTracer = otel.Tracer("AccountInquiriesRepository")

type AccountInquiriesRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IAccountInquiriesRepository {
	return &AccountInquiriesRepository{
		db:     db,
		logger: logger,
	}
}
