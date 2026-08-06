package userRole

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("UserRoleService")

type UserRoleService struct {
	repo   repository.IUserRoleRepository
	logger logger.ILogger
}

func New(repo repository.IUserRoleRepository, logger logger.ILogger) service.IUserRoleService {
	return &UserRoleService{
		repo:   repo,
		logger: logger,
	}
}
