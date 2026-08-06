package user

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserRepository")

type UserRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

const tableName = "users"

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IUserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}
