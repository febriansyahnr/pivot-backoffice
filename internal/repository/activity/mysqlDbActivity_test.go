package activityRepository_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	activityModel "github.com/paper-indonesia/pivot-backoffice/internal/model/activity"
	activityRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/activity"
	mongoDbMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mongoDbExt"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOnMySQL(t *testing.T) {
	activity := activityModel.Activity{
		ID:          uuid.NewString(),
		MerchantID:  uuid.NewString(),
		UserID:      &userID,
		Tag:         constant.TagAccount,
		Activity:    constant.ActivityUserLogin,
		ServiceName: "DoLogin",
		Parameter: &map[string]any{
			"email":    "jay@paper.id",
			"password": "*****",
		},
		CreatedAt: util.TimeNow,
		UpdatedAt: util.TimeNow,
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Insert Activity Logs to MySQL",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*activityModel.ActivityDTO"),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Failure insert Activity Logs to MySQL",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*activityModel.ActivityDTO"),
				).Return(false, errors.New("error when inserting activity_logs"))
			},
			wantErr: true,
		},
		{
			name: "FAILED: No rows inserted in Activity Logs to MySQL",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"NamedExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("*activityModel.ActivityDTO"),
				).Return(false, nil)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMongoDb := mongoDbMocks.NewIMongoDbExt(t)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			factory := activityRepository.ActivityRepository{
				Mongo:  mockMongoDb,
				Mysql:  mockMysql,
				Logger: mockLogger,
			}
			repo := factory.CreateRepository(activityRepository.MySQLType)
			ctx := context.Background()
			err := repo.Create(ctx, &activity)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMongoDb.AssertExpectations(t)
		})
	}
}

func TestGetListMySQL(t *testing.T) {
	merchantID := uuid.NewString()
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows)
		filter    activityModel.ActivityFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with merchantID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: activityModel.ActivityFilterRequest{
				MerchantID: &merchantID,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with created_at filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: activityModel.ActivityFilterRequest{
				MerchantID:     &merchantID,
				StartCreatedAt: &util.TimeNow,
				EndCreatedAt:   &util.TimeNow,
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("invalid table name"))
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: true,
		},
		{
			name: "Get List on error retrieving data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("invalid scan data"))

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)
			},
			filter:  activityModel.ActivityFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMongoDb := mongoDbMocks.NewIMongoDbExt(t)
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysqlRows := mysqlMocks.NewIMySqlRows(t)
			tc.mockSetup(mockMysql, mockMysqlRows)

			factory := activityRepository.ActivityRepository{
				Mongo:  mockMongoDb,
				Mysql:  mockMysql,
				Logger: mockLogger,
			}
			repo := factory.CreateRepository(activityRepository.MySQLType)
			ctx := context.Background()
			_, err := repo.GetList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMongoDb.AssertExpectations(t)
		})
	}
}

func TestFindLastMerchantActivityDate(t *testing.T) {
	timeNow := time.Now()

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					existedPtr := args.Get(1).(*time.Time)
					*existedPtr = timeNow
				}).Return(nil)
			},
		},
		{
			name: "ERROR: Mysql error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			wantErr: true,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.TimeReferenceMockType(), constant.StringMockType(), constant.StringMockType(),
				).Return(sql.ErrNoRows)
			},
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			ctx := context.Background()

			tc.mockSetup(mockMysql)

			factory := activityRepository.ActivityRepository{
				Mongo:  nil,
				Mysql:  mockMysql,
				Logger: mockLogger,
			}
			repo := factory.CreateRepository(activityRepository.MySQLType)

			if _, err := repo.FindLastMerchantActivityDate(ctx, uuid.NewString()); tc.wantErr {
				assert.Error(t, err)

			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}
