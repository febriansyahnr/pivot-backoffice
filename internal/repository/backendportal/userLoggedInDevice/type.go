package userLoggedInDeviceRepository

import (
	"github.com/paper-indonesia/pdk/v2/logger"
	repository "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserLoggedInDeviceRepository")

type UserLoggedInDeviceRepository struct {
	db     mySqlExt.IMySqlExt
	logger logger.ILogger
}

func New(db mySqlExt.IMySqlExt, logger logger.ILogger) repository.IUserLoggedInDeviceRepository {
	return &UserLoggedInDeviceRepository{
		db:     db,
		logger: logger,
	}
}
