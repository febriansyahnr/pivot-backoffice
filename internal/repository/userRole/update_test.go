package userRole

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/userRole"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestUserRoleRepository_Update(t *testing.T) {
	now := time.Now()

	existedRole := &userRole.UserRole{
		UUID:      "uuid-uuid-uuid",
		UserID:    "user-id",
		RoleID:    "role-id",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		role      *userRole.UserRole
		wantErr   bool
	}{
		{
			name: "Valid User",
			role: existedRole,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure Insert to Database",
			role: existedRole,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("update error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			err := repo.UpdateByUserID(context.Background(), tc.role)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.Update() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
