package role

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mockRole "github.com/paper-indonesia/pivot-backoffice/mocks/repository/role"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdate(t *testing.T) {
	now := time.Now()

	createdRole := &role.Role{
		UUID:      "uuid-uuid-uuid",
		Name:      "admin",
		Slug:      "admin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name       string
		input      *role.Role
		mocksSetup func(trxRepo *mockRole.IRoleRepository)
		wantErr    bool
	}{
		{
			name:  "SUCCESS: successfully update role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockRole.IRoleRepository) {
				trxRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*role.Role"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name:  "ERROR: error update role",
			input: createdRole,
			mocksSetup: func(trxRepo *mockRole.IRoleRepository) {
				trxRepo.On(
					"Update",
					mock.Anything,
					mock.AnythingOfType("*role.Role"),
				).Return(errors.New("insert error"))
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
			err := trxSvc.Update(ctx, tc.input)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			roleRepo.AssertExpectations(t)
		})
	}
}
