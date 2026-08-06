package role

import (
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

// List of menu combinations that are not allowed
//
// Key: List of menu (menu slug)
var combinationsAreNotAllowed = map[string][]string{
	"disbursement-transaction": {"disbursement.disbursement-create", "disbursement.disbursement-approval"},
}
var otelTracer = otel.Tracer("RoleService")

type RoleService struct {
	repo             repository.IRoleRepository
	logger           logger.ILogger
	menuRepo         repository.IMenuRepository
	roleMenuPermRepo repository.IRoleMenuPermissionRepository
	userRoleRepo     repository.IUserRoleRepository
	redis            redisExt.IRedisExt
}

type DependFunc func(*RoleService)

func New(repo repository.IRoleRepository, logger logger.ILogger, depends ...DependFunc) service.IRoleService {
	r := &RoleService{
		repo:   repo,
		logger: logger,
	}
	for _, f := range depends {
		f(r)
	}
	return r
}

func WithMenuRepository(repo repository.IMenuRepository) DependFunc {
	return func(r *RoleService) {
		r.menuRepo = repo
	}
}

func WithRoleMenuPermissionRepository(repo repository.IRoleMenuPermissionRepository) DependFunc {
	return func(r *RoleService) {
		r.roleMenuPermRepo = repo
	}
}

func WithUserRoleRepository(repo repository.IUserRoleRepository) DependFunc {
	return func(r *RoleService) {
		r.userRoleRepo = repo
	}
}

func WithRedisClient(rdb redisExt.IRedisExt) DependFunc {
	return func(ds *RoleService) {
		ds.redis = rdb
	}
}
