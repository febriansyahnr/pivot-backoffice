package role

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFindRoleBySlug(t *testing.T) {
	now := time.Now()

	expectedRole := &role.Role{
		UUID:      "uuid-uuid-uuid",
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		Name           string
		IsSuccess      bool
		RoleID         string
		ExpectedResult *role.RoleResponse
		ExpectedError  string
		MockSetup      func(mockRepo *mockRole.IRoleRepository)
	}{
		{
			Name:           "SUCCESS: find role by slug",
			IsSuccess:      true,
			RoleID:         "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
			ExpectedResult: expectedRole.ToResponse(),
			MockSetup: func(mockRepo *mockRole.IRoleRepository) {
				mockRepo.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(expectedRole, nil)
			},
		},
		{
			Name:          "ERROR: role not found",
			IsSuccess:     false,
			RoleID:        "not-found",
			ExpectedError: "ERROR_NOT_FOUND",
			MockSetup: func(mockRepo *mockRole.IRoleRepository) {
				mockRepo.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
		},
		{
			Name:          "ERROR: error find role",
			IsSuccess:     false,
			RoleID:        "user-error",
			ExpectedError: "error find role",
			MockSetup: func(mockRepo *mockRole.IRoleRepository) {
				mockRepo.On(
					"FindRoleBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("error find role"))
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			roleRepo := mockRole.NewIRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.MockSetup(roleRepo)

			svc := New(roleRepo, loggerMock)

			response, err := svc.FindRoleBySlug(context.Background(), tc.RoleID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			}

			roleRepo.AssertExpectations(t)
		})
	}
}
