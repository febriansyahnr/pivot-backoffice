package role

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindRoleById(t *testing.T) {
	testCases := []struct {
		name       string
		mocksSetup func(roleRepo *mockRole.IRoleRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully find role",
			mocksSetup: func(roleRepo *mockRole.IRoleRepository) {
				roleRepo.On(
					"FindRoleByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(&role.Role{}, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: failed to find role",
			mocksSetup: func(roleRepo *mockRole.IRoleRepository) {
				roleRepo.On(
					"FindRoleByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: role not found",
			mocksSetup: func(roleRepo *mockRole.IRoleRepository) {
				roleRepo.On(
					"FindRoleByID",
					mock.Anything,
					mock.AnythingOfType("string"),
				).Return(nil, nil)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleRepo := mockRole.NewIRoleRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})

			tc.mocksSetup(roleRepo)

			trxSvc := New(roleRepo, loggerMock)

			ctx := context.Background()
			_, err := trxSvc.FindRoleById(ctx, uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			roleRepo.AssertExpectations(t)
		})
	}
}
