package requestaccountinquiry

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

const tableName = "request_account_inquiries"

var otelTracer = otel.Tracer("RequestAccountInquiryRepository")

type RequestAccountInquiryRepository struct {
	*basicsql.Properties

	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IRequestAccountInquiryRepository {
	return &RequestAccountInquiryRepository{
		Properties: basicsql.NewBasicSQLProperties(db),
		db:         db,
		logger:     logger,
	}
}
