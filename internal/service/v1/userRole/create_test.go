package userRole

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/userRole"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserRoleService_Create(t *testing.T) {
	now := time.Now()

	createdRole := &userRole.UserRole{
		UUID:      "uuid-uuid-uuid",
		UserID:    "user-id",
		RoleID:    "role-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name       string
		input      *userRole.UserRole
		mocksSetup func(trxRepo *mockUserRole.IUserRoleRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully create user_role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockUserRole.IUserRoleRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*userRole.UserRole"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: error create role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockUserRole.IUserRoleRepository) {
				trxRepo.On(
					"Create",
					mock.Anything,
					mock.AnythingOfType("*userRole.UserRole"),
				).Return(errors.New("insert error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userRoleRepo := mockUserRole.NewIUserRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(userRoleRepo)

			trxSvc := New(userRoleRepo, loggerMock)

			err := trxSvc.Create(context.Background(), tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			userRoleRepo.AssertExpectations(t)
		})
	}
}
