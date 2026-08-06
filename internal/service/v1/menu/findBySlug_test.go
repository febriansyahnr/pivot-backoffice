package menuService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
)

func TestFindBySlug(t *testing.T) {
	testCases := []struct {
		name       string
		input      *role.Role
		mocksSetup func(menuRepo *repositoryMocks.IMenuRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"FindBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&menuModel.Menu{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: FindBySlug service",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"FindBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: FindBySlug not found",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"FindBySlug",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			menuRepo := repositoryMocks.NewIMenuRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(menuRepo)

			svc := New(menuRepo, loggerMock)

			ctx := context.Background()
			_, err := svc.FindBySlug(ctx, "slug")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			menuRepo.AssertExpectations(t)

		})
	}
}
