package permissionService

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	permissionModel "github.com/paper-indonesia/pivot-backoffice/internal/model/permission"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
)

func TestGetCachedPermissionsByRoleId(t *testing.T) {
	permissionRepo := repositoryMocks.NewIPermissionRepository(t)
	loggerMock, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	db, r := redismock.NewClientMock()
	redisMock := redisExt.WrapRedisClient(db, nil)

	roleId := uuid.NewString()
	cacheKey := fmt.Sprintf(constant.PermissionByRoleKeyPattern, roleId)

	tests := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Get cache",
			wantErr: true,
			setupMock: func() {
				r.ExpectGet(cacheKey).SetErr(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: Get from cache",
			wantErr: false,
			setupMock: func() {
				r.ExpectGet(cacheKey).SetVal(strings.Join([]string{constant.PermissionSlugDisbursementView}, ","))
			},
		},
		{
			name:    "ERROR: Find permissions by role ID",
			wantErr: true,
			setupMock: func() {
				r.ExpectGet(cacheKey).SetErr(redis.Nil)

				permissionRepo.On("FindByRoleId", constant.ValueCtxMockType(), roleId).Once().
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS: Get from DB and set cache",
			wantErr: false,
			setupMock: func() {
				r.ExpectGet(cacheKey).SetErr(redis.Nil)

				permissionRepo.On("FindByRoleId", constant.ValueCtxMockType(), roleId).
					Return([]*permissionModel.Permission{{Slug: constant.PermissionSlugDisbursementView}}, nil)

				r.ExpectHSet(cacheKey, strings.Join([]string{constant.PermissionSlugDisbursementView}, ",")).SetVal(1)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			svc := New(permissionRepo, loggerMock, WithRedisClient(redisMock))
			_, err := svc.GetCachedPermissionsByRoleId(context.Background(), roleId)
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			permissionRepo.AssertExpectations(t)
		})
	}
}
