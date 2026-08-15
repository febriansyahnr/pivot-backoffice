package ipWhitelistRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	ipwhitelistModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/ipWhitelist"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetList(t *testing.T) {

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"Rebind",
					constant.StringMockType(),
				).Return("")

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Get List",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"Rebind",
					constant.StringMockType(),
				).Return("")

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(errors.New("errors"))

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Count",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"Rebind",
					constant.StringMockType(),
				).Return("")

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(errors.New("errors"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: No rows",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"Rebind",
					constant.StringMockType(),
				).Return("")

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(sql.ErrNoRows)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)

			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Without ExcludedIDs",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"Rebind",
					constant.StringMockType(),
				).Return("")

				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.Int64MockType(),
					constant.Int64MockType(),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: sqlx.In fails during ExcludedIDs processing",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// No DB mocks needed as error occurs before DB calls
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockLogger, mockMysql).(*IPWhitelistRepository)

			// Inject mock sqlxIn function for error test case
			if tc.name == "ERROR: sqlx.In fails during ExcludedIDs processing" {
				repo.sqlxIn = func(query string, args ...interface{}) (string, []interface{}, error) {
					return "", nil, errors.New("mocked sqlx.In error")
				}
			}

			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, IPWhitelistTable)

			req := &ipwhitelistModel.GetIPWhitelistConfiguration{
				MerchantID: uuid.NewString(),
				IP:         "123.123.123.123",
				Subnet:     "24",
				Status:     "ACTIVE",
				Page:       1,
				PageSize:   10,
			}

			// Configure ExcludedIDs based on test case
			switch tc.name {
			case "SUCCESS: Without ExcludedIDs":
				// Leave ExcludedIDs as nil
			case "ERROR: sqlx.In fails during ExcludedIDs processing":
				// Add ExcludedIDs to trigger the sqlxIn call
				req.ExcludedIDs = []string{uuid.NewString()}
			default:
				// Add valid ExcludedIDs for other test cases
				req.ExcludedIDs = []string{uuid.NewString()}
			}

			_, _, err := repo.List(ctx, req)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
