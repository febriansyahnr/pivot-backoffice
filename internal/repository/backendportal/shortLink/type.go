package shortLinkRepository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository/basicsql"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer = otel.Tracer("ShortLinkRepository")
	tableName  = "short_links"
)

type shortLinkRepo struct {
	*basicsql.Properties

	db  mySqlExt.IMySqlExt
	log logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IShortLinkRepository {
	return &shortLinkRepo{basicsql.NewBasicSQLProperties(db), db, logger}
}
