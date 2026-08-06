package role

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetList(t *testing.T) {
	data := make([]roleModel.Role, 0)
	data = append(data, roleModel.Role{
		UUID:      "",
		Name:      "",
		Slug:      "",
		Type:      "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	expectedResponse := commonModel.PaginationResponse{
		Data: data,
		Meta: commonModel.Meta{
			Page:       1,
			PerPage:    20,
			TotalItems: 1,
			TotalPages: 1,
		},
	}

	testCases := []struct {
		name           string
		WantErr        bool
		ExpectedResult *commonModel.PaginationResponse
		ExpectedError  string
		mockSetup      func(mockRepo *mockRole.IRoleRepository)
	}{
		{
			name:           "SUCCESS: list roles",
			WantErr:        false,
			ExpectedResult: &expectedResponse,
			mockSetup: func(mockRepo *mockRole.IRoleRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.GetRoleFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(&expectedResponse, nil)
			},
		},
		{
			name:           "ERROR: error list roles",
			WantErr:        true,
			ExpectedResult: &expectedResponse,
			ExpectedError:  "ERROR_DATABASE | error list roles",
			mockSetup: func(mockRepo *mockRole.IRoleRepository) {
				mockRepo.On(
					"GetList",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*role.GetRoleFilterRequest"),
					mock.AnythingOfType("int64"),
					mock.AnythingOfType("int64"),
				).Return(nil, fmt.Errorf("error list roles"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			roleMock := mockRole.NewIRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			ctx := context.Background()

			tc.mockSetup(roleMock)

			svc := New(roleMock, loggerMock)

			results, err := svc.GetList(ctx, &roleModel.GetRoleFilterRequest{}, 1, 10)

			if tc.WantErr {
				require.Error(t, err)
				require.Empty(t, results)
				require.True(t, strings.Contains(err.Error(), tc.ExpectedError))
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, results)
			}

			roleMock.AssertExpectations(t)
		})
	}
}
