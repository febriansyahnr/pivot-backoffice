package callbackRepository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/callback"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/callback"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const ArrCallbackLogWithMaster = "*[]callback_model.CallbackLogWithMaster"

func TestGetCallbackLogList(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		filter    *callbackModel.GetListCallbackLogFilterRequest
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(ArrCallbackLogWithMaster),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter:  &callbackModel.GetListCallbackLogFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(ArrCallbackLogWithMaster),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(errors.New("no rows data"))
			},
			filter:  &callbackModel.GetListCallbackLogFilterRequest{},
			wantErr: false,
		},
		{
			name: "SUCCESS:  Get List with filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(ArrCallbackLogWithMaster),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType(constant.MockTypeTimeReference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: &callbackModel.GetListCallbackLogFilterRequest{
				MerchantID:     uuid.NewString(),
				Type:           constant.CallbackNameDisbursement,
				Event:          constant.CallbackEventPayoutDone,
				StartUpdatedAt: &util.TimeNow,
				EndUpdatedAt:   &util.TimeNow,
				Status:         constant.CallbackStatusDelivered,
				Keyword:        "key",
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with keyword filter (searches both UUID and reference_id)",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(ArrCallbackLogWithMaster),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
					mock.AnythingOfType("string"),
				).Return(nil)
			},
			filter: &callbackModel.GetListCallbackLogFilterRequest{
				MerchantID: uuid.NewString(),
				Keyword:    "test-reference-123",
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(ArrCallbackLogWithMaster),
					mock.AnythingOfType("string"),
				).Return(constant.ErrSomeErrorForUnitTest)

				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType(constant.MockTypeInt64Reference),
					mock.AnythingOfType("string"),
				).Return(nil)

			},
			filter:  &callbackModel.GetListCallbackLogFilterRequest{},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetCallbackLogList(ctx, tc.filter, 0, 20)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestFindMerchantCallbackHistory(t *testing.T) {
	log := pdkLogMock.NewILogger(t)
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, log)

	filter := callbackModel.GetListCallbackLogFilterRequest{
		MerchantID:     "0435dbe1-9283-432d-b611-93b65245ae6d",
		StartUpdatedAt: util.ValueToPtr(time.Date(2025, 12, 28, 17, 0, 0, 0, time.UTC)),
		EndUpdatedAt:   util.ValueToPtr(time.Date(2025, 12, 29, 16, 59, 59, 999, time.UTC)),
	}
	basicArgs := []any{filter.StartUpdatedAt, filter.EndUpdatedAt, filter.StartUpdatedAt, filter.EndUpdatedAt, filter.MerchantID}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *commonModel.PaginationResponse
	}{
		{
			name: "ERROR:List merchant history",
			setupMock: func() {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, basicArgs...)

				db.On("SelectContext", args...).Once().Run(func(mock.Arguments) { time.Sleep(30 * time.Millisecond) }).Return(assert.AnError)
				db.On("GetContext", args...).Once().Run(func(mock.Arguments) { time.Sleep(30 * time.Millisecond) }).Return(nil)
				log.On(
					"Error", mock.Anything, "Failed while find merchant callback history", mock.Anything,
				).Once().Return()
			},
			wantError: fmt.Errorf("list callback history: %v", assert.AnError),
		},
		{
			name: "ERROR:Calculate total rows",
			setupMock: func() {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, basicArgs...)

				db.On("SelectContext", args...).Once().Run(func(mock.Arguments) { time.Sleep(30 * time.Millisecond) }).Return(nil)
				db.On("GetContext", args...).Once().Run(func(mock.Arguments) { time.Sleep(30 * time.Millisecond) }).Return(assert.AnError)
				log.On(
					"Error", mock.Anything, "Failed while find merchant callback history", mock.Anything,
				).Once().Return()
			},
			wantError: fmt.Errorf("calculate total items: %v", assert.AnError),
		},
		{
			name: "SUCCESS:Without additional filter",
			setupMock: func() {
				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, basicArgs...)

				db.On("SelectContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*[]callbackModel.CallbackLogWithMaster) = []callbackModel.CallbackLogWithMaster{
						{UUID: util.ParseUUID("59fb22a2-ac73-4718-b156-b8b56828d482")},
					}
				}).Return(nil)
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*int64) = 1
				}).Return(nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []callbackModel.CallbackLogWithMaster{
					{UUID: util.ParseUUID("59fb22a2-ac73-4718-b156-b8b56828d482")},
				},
				Meta: commonModel.Meta{
					Page: 1, PerPage: 10, TotalItems: 1, TotalPages: 1,
				},
			},
		},
		{
			name: "SUCCESS:With additional filter",
			setupMock: func() {
				filter.Type = "PAYOUT"
				filter.Event = "PAYOUT.DONE"
				filter.Status = "DELIVERED"
				filter.Keyword = "TEST"

				args := append([]any{mock.Anything, mock.Anything, mock.Anything}, append(basicArgs, filter.Type, filter.Event, filter.Status, "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")...)

				db.On("SelectContext", args...).Once().Return(sql.ErrNoRows)
				db.On("GetContext", args...).Once().Return(nil)
			},
			wantResult: &commonModel.PaginationResponse{
				Data: []callbackModel.CallbackLogWithMaster{},
				Meta: commonModel.Meta{
					Page: 1, PerPage: 10, TotalItems: 0, TotalPages: 0,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.FindMerchantCallbackHistory(t.Context(), &filter, 1, 10)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
			log.AssertExpectations(t)
		})
	}
}
