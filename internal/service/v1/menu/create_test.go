package menuService

import (
	"context"
	"testing"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
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
					"Create",
					mock.Anything,
					mock.AnythingOfType("*menuModel.Menu"),
				).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			menuRepo := repositoryMocks.NewIMenuRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(menuRepo)

			svc := New(menuRepo, loggerMock)

			ctx := context.Background()
			err := svc.Create(ctx, &menuModel.Menu{})
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			menuRepo.AssertExpectations(t)

		})
	}
}
