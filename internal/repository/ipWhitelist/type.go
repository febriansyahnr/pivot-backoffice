package ipWhitelistRepository

import (
	"github.com/jmoiron/sqlx"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var (
	otelTracer       = otel.Tracer("IPWhitelistRepository")
	IPWhitelistTable = "ip_whitelist_configurations"
)

// SqlxInFunc is a function type for sqlx.In to allow dependency injection for testing
type SqlxInFunc func(query string, args ...interface{}) (string, []interface{}, error)

type IPWhitelistRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
	sqlxIn SqlxInFunc
}

func New(
	logger logger.ILogger,
	db mySqlExt.IMySqlExt,
) repository.IIPWhitelistRepository {
	return &IPWhitelistRepository{
		db:     db,
		logger: logger,
		sqlxIn: sqlx.In, // Default to real sqlx.In
	}
}
