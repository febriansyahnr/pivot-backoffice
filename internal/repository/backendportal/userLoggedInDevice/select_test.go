package userLoggedInDeviceRepository

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userLoggedInDeviceModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/userLoggedInDevice"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllByUserID(t *testing.T) {
	userLoggedInResult := []*userLoggedInDeviceModel.UserLoggedInDevice{
		{
			UUID:             uuid.NewString(),
			UserID:           uuid.NewString(),
			DeviceIdentifier: "tes-identifier",
		},
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  []*userLoggedInDeviceModel.UserLoggedInDevice
		wantErr   bool
	}{
		{
			name: "SUCCESS: GetAllByUserID",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*[]*userLoggedInDeviceModel.UserLoggedInDevice)
					*ptr = userLoggedInResult
				})
			},
			expected: userLoggedInResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Database GetAllByUserID Error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")
			transaction, err := repo.GetAllByUserID(ctx, uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindByUserAndDevice(t *testing.T) {
	userLoggedInResult := userLoggedInDeviceModel.UserLoggedInDevice{
		UUID:             uuid.NewString(),
		UserID:           uuid.NewString(),
		DeviceIdentifier: "tes-identifier",
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *userLoggedInDeviceModel.UserLoggedInDevice
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrUserLoggedInDeviceMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					ptr := args.Get(1).(*userLoggedInDeviceModel.UserLoggedInDevice)
					*ptr = userLoggedInResult
				})
			},
			expected: &userLoggedInResult,
			wantErr:  false,
		},
		{
			name: "ERROR: Not found data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrUserLoggedInDeviceMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
			expected: nil,
			wantErr:  false,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrUserLoggedInDeviceMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)
			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "user_logged_in_devices")
			transaction, err := repo.FindByUserAndDevice(ctx, uuid.NewString(), "device-id")
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
