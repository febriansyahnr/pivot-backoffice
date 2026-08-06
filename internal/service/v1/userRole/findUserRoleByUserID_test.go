package userRole

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mockUserRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/userRole"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserRoleService_FindUserRoleByUserID(t *testing.T) {
	now := time.Now()

	expectedRole := &userRole.UserRole{
		UUID:      "uuid-uuid-uuid",
		UserID:    "user-id",
		RoleID:    "role-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		Name           string
		IsSuccess      bool
		ID             string
		ExpectedResult *userRole.UserRole
		ExpectedError  string
		MockSetup      func(mockRepo *mockUserRole.IUserRoleRepository)
	}{
		{
			Name:           "SUCCESS: find user_role by user_id",
			IsSuccess:      true,
			ID:             "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
			ExpectedResult: expectedRole,
			MockSetup: func(mockRepo *mockUserRole.IUserRoleRepository) {
				mockRepo.On(
					"FindUserRoleByUserID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedRole, nil)
			},
		},
		{
			Name:          "ERROR: user_role not found",
			IsSuccess:     false,
			ID:            "not-found",
			ExpectedError: "ERROR_NOT_FOUND",
			MockSetup: func(mockRepo *mockUserRole.IUserRoleRepository) {
				mockRepo.On(
					"FindUserRoleByUserID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
		},
		{
			Name:          "ERROR: error find user_role",
			IsSuccess:     false,
			ID:            "user-role-error",
			ExpectedError: "error find user_role",
			MockSetup: func(mockRepo *mockUserRole.IUserRoleRepository) {
				mockRepo.On(
					"FindUserRoleByUserID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error find user_role"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			userRoleRepo := mockUserRole.NewIUserRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(userRoleRepo)

			svc := New(userRoleRepo, loggerMock)

			response, err := svc.FindUserRoleByUserID(context.Background(), tc.ID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			userRoleRepo.AssertExpectations(t)
		})
	}
}
