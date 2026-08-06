package permissionService

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
)

const cacheDuration = 24 * time.Hour

func (s *PermissionService) GetCachedPermissionsByRoleId(ctx context.Context, roleId string) ([]string, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/permission/GetCachedPermissionsByRoleId")
	defer segment.End()

	cacheKey := fmt.Sprintf(constant.PermissionByRoleKeyPattern, roleId)

	// If key not exist in redis, then get from DB
	if permissionInStr, err := s.redis.Get(ctx, cacheKey).Result(); err != nil && err != redis.Nil {
		s.logger.Error(ctx, "Error getting value from Redis", logger.String("roleId", roleId), logger.Error(err))
		return []string{}, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
	} else if permissionInStr != "" {
		return strings.Split(permissionInStr, ","), nil
	}

	// Get permissions by RoleId
	permissions, err := s.FindByRoleId(ctx, roleId)
	if err != nil {
		return nil, err
	}

	permissionInArr := make([]string, len(permissions))
	for i, permission := range permissions {
		permissionInArr[i] = permission.Slug
	}

	s.redis.Set(ctx, cacheKey, strings.Join(permissionInArr, ","), cacheDuration)

	return permissionInArr, nil
}
