package menuService

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	menuModel "github.com/paper-indonesia/pivot-backoffice/internal/model/menu"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByRole(t *testing.T) {
	parentIDs := []string{uuid.NewString(), uuid.NewString()}
	menuResponse := []*menuModel.MenuResponse{
		{
			UUID:      parentIDs[0],
			Slug:      "test",
			Name:      "name",
			Icon:      "icon",
			Path:      "/menu",
			Level:     1,
			Order:     1,
			ParentID:  nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Permissions: []menuModel.MenuPermission{
				{
					Group: "group",
					Slug:  "slug",
				},
			},
			Children: []*menuModel.MenuResponse{
				{
					UUID:      uuid.NewString(),
					Slug:      "test",
					Name:      "name",
					Icon:      "icon",
					Path:      "/menu",
					Level:     1,
					Order:     1,
					ParentID:  &parentIDs[0],
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Permissions: []menuModel.MenuPermission{
						{
							Group: "group",
							Slug:  "slug",
						},
					},
				},
			},
			AllowedProducts: &[]string{"TEST"},
		},
		{
			UUID:      parentIDs[1],
			Slug:      "test",
			Name:      "name",
			Icon:      "icon",
			Path:      "/menu",
			Level:     1,
			Order:     1,
			ParentID:  &parentIDs[0],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			Permissions: []menuModel.MenuPermission{
				{
					Group: "group",
					Slug:  "slug",
				},
			},
			Children: []*menuModel.MenuResponse{
				{
					UUID:      uuid.NewString(),
					Slug:      "test",
					Name:      "name",
					Icon:      "icon",
					Path:      "/menu",
					Level:     1,
					Order:     1,
					ParentID:  &parentIDs[1],
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Permissions: []menuModel.MenuPermission{
						{
							Group: "group",
							Slug:  "slug",
						},
					},
				},
			},
		},
	}

	disbursementPermission := []menuModel.MenuPermission{
		{
			Group: "Disbursement",
			Slug:  "disbursement.view",
		},
	}

	menuResponseWithDisbursement := []*menuModel.MenuResponse{
		{
			UUID:        parentIDs[0],
			Slug:        "disbursement",
			Name:        "Disbursement",
			Icon:        "disbursement.svg",
			Path:        "/disbursement",
			Level:       1,
			Order:       1,
			ParentID:    nil,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Permissions: disbursementPermission,
		},
	}

	homeMenu := &menuModel.MenuResponse{
		UUID:      uuid.NewString(),
		Slug:      "home",
		Name:      "Home",
		Icon:      "home.svg",
		Path:      "/home",
		Level:     0,
		Order:     1,
		ParentID:  nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Permissions: []menuModel.MenuPermission{
			{
				Group: "Home",
				Slug:  "home.view",
			},
		},
	}

	testCases := []struct {
		name       string
		input      *role.Role
		mocksSetup func(menuRepo *repositoryMocks.IMenuRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS - without Home menu (no disbursement permissions)",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(menuResponse, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS - with Home menu (has disbursement permissions)",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(menuResponseWithDisbursement, nil)
				menuRepo.On(
					"FindBySlugWithPermissions",
					mock.Anything,
					"home",
				).Return(homeMenu, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: GetAll service error",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS - FindBySlugWithPermissions error is logged and gracefully handled",
			mocksSetup: func(menuRepo *repositoryMocks.IMenuRepository) {
				menuRepo.On(
					"GetAll",
					mock.Anything,
					mock.AnythingOfType("*menuModel.GetAllFilterRequest"),
				).Return(menuResponseWithDisbursement, nil)
				menuRepo.On(
					"FindBySlugWithPermissions",
					mock.Anything,
					"home",
				).Return(nil, constant.ErrSomeErrorForUnitTest)
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
			_, err := svc.GetByRole(ctx, uuid.NewString(), true)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			menuRepo.AssertExpectations(t)

		})
	}
}
