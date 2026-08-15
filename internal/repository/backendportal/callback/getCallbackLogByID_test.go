package callbackRepository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	callbackRepository "github.com/paper-indonesia/pivot-backoffice/internal/repository/callback"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCallbackLogByID(t *testing.T) {
	eventName := constant.CallbackEventPayoutDone
	callbackLogWithMaster := callbackModel.CallbackLogWithMaster{
		UUID:       uuid.New(),
		CallbackID: uuid.New(),
		Type:       constant.CallbackNameDisbursement,
		BaseURL:    nil,
		URL:        "http://localhost",
		Event:      &eventName,
		Request:    "{}",
		Response:   nil,
		Status:     constant.CallbackStatusDelivered,
		Retry:      0,
		CreatedAt:  util.TimeNow,
		UpdatedAt:  util.TimeNow,
	}

	testCases := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		expected  *callbackModel.CallbackLogWithMaster
		wantErr   bool
	}{
		{
			name: "SUCCESS",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrCallbackLogWithMasterMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*callbackModel.CallbackLogWithMaster) = callbackLogWithMaster
				})
			},
			expected: &callbackLogWithMaster,
			wantErr:  false,
		},
		{
			name: "ERROR: Data not found",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					constant.ValueCtxMockType(),
					constant.PtrCallbackLogWithMasterMockType(),
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
					constant.PtrCallbackLogWithMasterMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(constant.ErrSomeErrorForUnitTest)

			},
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tt.mockSetup(mockMysql)

			repo := callbackRepository.New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, callbackRepository.TableCallbackLog)
			transaction, err := repo.GetCallbackLogByID(ctx, uuid.NewString())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, transaction)
			}
			mockMysql.AssertExpectations(t)

		})
	}
}
