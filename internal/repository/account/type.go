package account_repository

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/tracer"

	"github.com/paper-indonesia/pdk/v2/logger"
)

var otelTracer = tracer.New("AccountRepository")

const (
	TableName         = "accounts"
	MerchantTableName = "merchants"
	CustomerTableName = "customers"
)

type AccountRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IAccountRepository {
	return &AccountRepository{
		db:     db,
		logger: logger,
	}
}
