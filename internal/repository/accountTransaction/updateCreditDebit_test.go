package accounttransaction_repository

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCreditDebitByID(t *testing.T) {
	credit50 := 50.0
	debit30 := 30.0

	testCases := []struct {
		name      string
		id        string
		credit    *float64
		debit     *float64
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		errMsg    string
	}{
		{
			name:   "SUCCESS: Update with credit only",
			id:     "test-uuid-1",
			credit: &credit50,
			debit:  nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Update with debit only",
			id:     "test-uuid-2",
			credit: nil,
			debit:  &debit30,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Update with both credit and debit",
			id:     "test-uuid-3",
			credit: &credit50,
			debit:  &debit30,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name:      "ERROR: Neither credit nor debit provided",
			id:        "test-uuid-4",
			credit:    nil,
			debit:     nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// No mock setup needed as error is returned before DB call
			},
			wantErr: true,
			errMsg:  "either credit or debit value must be provided",
		},
		{
			name:   "ERROR: Database error",
			id:     "test-uuid-5",
			credit: &credit50,
			debit:  nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, errors.New("database connection failed"))
			},
			wantErr: true,
			errMsg:  "database connection failed",
		},
		{
			name:   "ERROR: No rows affected - data not found",
			id:     "test-uuid-6",
			credit: &credit50,
			debit:  nil,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, nil)
			},
			wantErr: true,
			errMsg:  "data not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, tableName)

			err := repo.UpdateCreditDebitByID(ctx, tc.id, tc.credit, tc.debit)

			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMsg != "" {
					assert.Contains(t, err.Error(), tc.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}
