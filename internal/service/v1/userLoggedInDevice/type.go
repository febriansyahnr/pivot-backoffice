package userLoggedInDeviceService

import (
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserLoggedInDeviceService")

type UserLoggedInDeviceService struct {
	config                 *config.Config
	secret                 *config.Secret
	logger                 logger.ILogger
	userLoggedInDeviceRepo repository.IUserLoggedInDeviceRepository
	userRepo               repository.IUserRepository
}

type Repositories struct {
	UserLoggedInDeviceRepo repository.IUserLoggedInDeviceRepository
	UserRepo               repository.IUserRepository
}

func New(
	config *config.Config,
	secret *config.Secret,
	logger logger.ILogger,
	repo Repositories,
) service.IUserLoggedInDeviceService {
	return &UserLoggedInDeviceService{
		config:                 config,
		secret:                 secret,
		logger:                 logger,
		userLoggedInDeviceRepo: repo.UserLoggedInDeviceRepo,
		userRepo:               repo.UserRepo,
	}
}
