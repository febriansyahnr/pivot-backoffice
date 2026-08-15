package bankAccountRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"go.opentelemetry.io/otel"
)

type bankAccountRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

var (
	tableName  = "bank_accounts"
	otelTracer = otel.Tracer("BankAccountRepository")
)

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IBankAccountRepository {
	return &bankAccountRepository{db, logger}
}
