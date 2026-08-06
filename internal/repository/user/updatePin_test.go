package user

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestUpdatePin(t *testing.T) {
	// Define the test cases
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		userId    string
		hashedPin string
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update pin in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(true, nil)
			},
			wantErr:   false,
			userId:    user.UUID,
			hashedPin: "hashed pin",
		},
		{
			name: "FAILED: Error when updating pin in user account",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("error database"))
			},
			wantErr:   true,
			userId:    user.UUID,
			hashedPin: "hashed pin",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "users")
			err := repo.UpdatePin(ctx, tc.userId, tc.hashedPin)

			if (err != nil) != tc.wantErr {
				t.Errorf("UserRepository.UpdatePin() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
