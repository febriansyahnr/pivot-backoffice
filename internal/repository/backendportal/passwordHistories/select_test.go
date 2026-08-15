package passwordHistories

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/passwordHistories"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	passwordHistory = &passwordHistories.PasswordHistories{
		UUID:           "uuid-uuid-uuid",
		UserID:         "uuid-uuid-uuid",
		PasswordHashes: "password-hashes",
		CreatedAt:      time.Now(),
	}
)

func TestPasswordHistoriesRepository_FindUserByID(t *testing.T) {
	histories := []*passwordHistories.PasswordHistories{passwordHistory}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		input     string
		expected  []*passwordHistories.PasswordHistories
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Password Histories By User ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*passwordHistories.PasswordHistories"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					historiesPtr := args.Get(1).(*[]*passwordHistories.PasswordHistories)
					*historiesPtr = histories
				})
			},
			input:    "uuid-uuid-uuid",
			expected: histories,
			wantErr:  false,
		},
		{
			name: "ERROR: Find Password Histories By User ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*[]*passwordHistories.PasswordHistories"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(assert.AnError).Run(func(args mock.Arguments) {
					historiesPtr := args.Get(1).(*[]*passwordHistories.PasswordHistories)
					*historiesPtr = histories
				})
			},
			input:    "uuid-uuid-uuid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			limit := 10
			actual, err := repo.FindByUserID(context.Background(), tc.input, &limit)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestPasswordHistoriesRepository_FindByPassHashAndUserID(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		passHash  string
		userId    string
		expected  *passwordHistories.PasswordHistories
		wantErr   bool
	}{
		{
			name: "SUCCESS: Find Password Histories By Pass Hash And User ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*passwordHistories.PasswordHistories"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil).Run(func(args mock.Arguments) {
					historiesPtr := args.Get(1).(*passwordHistories.PasswordHistories)
					*historiesPtr = *passwordHistory
				})
			},
			passHash: "password-hashes",
			userId:   "uuid-uuid-uuid",
			expected: passwordHistory,
			wantErr:  false,
		},
		{
			name: "ERROR: Find Password Histories By Pass Hash And User ID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*passwordHistories.PasswordHistories"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(assert.AnError).Run(func(args mock.Arguments) {
					historiesPtr := args.Get(1).(*passwordHistories.PasswordHistories)
					*historiesPtr = *passwordHistory
				})
			},
			passHash: "password-hashes",
			userId:   "uuid-uuid-uuid",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			actual, err := repo.FindByPassHashAndUserID(context.Background(), tc.userId, tc.passHash)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
