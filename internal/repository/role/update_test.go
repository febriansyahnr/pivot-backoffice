package role

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	roleModel "github.com/paper-indonesia/pivot-backoffice/internal/model/role"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestRoleRepository_Update(t *testing.T) {
	expectedRole := &roleModel.Role{
		UUID:       "49426fa4-2f80-4b88-a8ae-39daf33d3e89",
		MerchantID: sql.NullString{},
		Name:       "a",
		Slug:       "a",
		Type:       "a",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		DeletedAt:  sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
	}

	// Define the test cases
	testCases := []struct {
		name      string
		inputRole *roleModel.Role
		result    error
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update role",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr:   false,
			inputRole: expectedRole,
			result:    nil,
		},
		{
			name: "FAILED: Error when updating blocked status in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("error database"))
			},
			wantErr:   true,
			inputRole: nil,
			result:    errors.New("error database"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "roles")
			err := repo.Update(ctx, expectedRole)

			if (err != nil) != tc.wantErr {
				t.Errorf("RoleRepository.UpdateBlocked() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
