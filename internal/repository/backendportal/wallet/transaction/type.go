package walletTransaction

import (
	port "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"go.opentelemetry.io/otel"
)

const tableName = "account_transactions"

var otelTracer = otel.Tracer("WalletTransactionRepository")

type repository struct {
	db mySqlExt.IMySqlExt
}

func New(db mySqlExt.IMySqlExt) port.IWalletTransactionRepository {
	return &repository{db}
}
