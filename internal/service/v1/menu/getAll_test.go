package menuService

import (
	"context"
	"testing"

	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAll(t *testing.T) {
	allMenus := []*menuModel.MenuResponse{
		{
			UUID: uuid.NewString(),
			Slug: "home",
			Name: "Home",
		},
		{
			UUID: uuid.NewString(),
			Slug: "disbursement",
			Name: "Disbursement",
		},
		{
			UUID: uuid.NewString(),
			Slug: "payment",
			Name: "Payment",
		},
	}

	testCases := []struct {
		name         string
		excludeHome  bool
		mocksSetup   func(menuRepo *repositoryMocks.IMenuRepository)
		wantErr      bool
		validateFunc func(t *testing.T, result []*menuModel.MenuResponse)
	}{
		{
			name:        "SUCCESS - Include all menus (excludeHome=false)",
			excludeHome: false,
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(allMenus, nil)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result []*menuModel.MenuResponse) {
				require.Len(t, result, 3)
				assert.Equal(t, "home", result[0].Slug)
				assert.Equal(t, "disbursement", result[1].Slug)
				assert.Equal(t, "payment", result[2].Slug)
			},
		},
		{
			name:        "SUCCESS - Exclude Home menu (excludeHome=true)",
			excludeHome: true,
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(allMenus, nil)
			},
			wantErr: false,
			validateFunc: func(t *testing.T, result []*menuModel.MenuResponse) {
				require.Len(t, result, 2)
				assert.Equal(t, "disbursement", result[0].Slug)
				assert.Equal(t, "payment", result[1].Slug)
				// Ensure Home menu is not included
				for _, menu := range result {
					assert.NotEqual(t, "home", menu.Slug)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			menuRepo := repositoryMocks.NewIMenuRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(menuRepo)

			svc := New(menuRepo, loggerMock)

			ctx := context.Background()
			result, err := svc.GetAll(ctx, tc.excludeHome)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tc.validateFunc != nil {
					tc.validateFunc(t, result)
				}
			}

			menuRepo.AssertExpectations(t)

		})
	}
}
