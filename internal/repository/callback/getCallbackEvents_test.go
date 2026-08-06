package callbackRepository_test

import (
	"context"
	"errors"
	"testing"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/callback"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const ArrCallbackEvent = "*[]callback_model.CallbackEvent"

func TestGetCallbackEvents(t *testing.T) {
	testCases := []struct {
		name       string
		mockSetup  func(mysqlMock *mysqlMocks.IMySqlExt, logMock *pdkLogMock.ILogger)
		wantErr    bool
		wantLength int
	}{
		{
			name: "SUCCESS: Get all active callback events",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, logMock *pdkLogMock.ILogger) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType(ArrCallbackEvent),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					events := args.Get(1).(*[]callbackModel.CallbackEvent)
					*events = []callbackModel.CallbackEvent{
						{
							UUID:       uuid.New(),
							Event:      "PAYOUT.DONE",
							Label:      "Payout Done",
							EventGroup: "Payout",
							IsActive:   true,
						},
						{
							UUID:       uuid.New(),
							Event:      "PAYMENT.VIRTUAL-ACCOUNT.PAID",
							Label:      "Payment Virtual Account Paid",
							EventGroup: "Payment",
							IsActive:   true,
						},
					}
				}).Return(nil)
			},
			wantErr:    false,
			wantLength: 2,
		},
		{
			name: "SUCCESS: Get empty callback events list",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, logMock *pdkLogMock.ILogger) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType(ArrCallbackEvent),
					mock.AnythingOfType("string"),
				).Run(func(args mock.Arguments) {
					events := args.Get(1).(*[]callbackModel.CallbackEvent)
					*events = []callbackModel.CallbackEvent{}
				}).Return(nil)
			},
			wantErr:    false,
			wantLength: 0,
		},
		{
			name: "ERROR: Database error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, logMock *pdkLogMock.ILogger) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.AnythingOfType(ArrCallbackEvent),
					mock.AnythingOfType("string"),
				).Return(errors.New("database connection error"))
				logMock.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger := pdkLogMock.NewILogger(t)
			tc.mockSetup(mockMysql, mockLogger)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			events, err := repo.GetCallbackEvents(ctx)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, events)
			} else {
				require.NoError(t, err)
				require.NotNil(t, events)
				require.Len(t, events, tc.wantLength)
			}

			mockMysql.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}

func TestGetCallbackEvents_WithData(t *testing.T) {
	expectedEvents := []callbackModel.CallbackEvent{
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Event:      "PAYOUT.DONE",
			Label:      "Payout Done",
			EventGroup: "Payout",
			IsActive:   true,
		},
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440001"),
			Event:      "PAYOUT.PENDING",
			Label:      "Payout Pending",
			EventGroup: "Payout",
			IsActive:   true,
		},
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Event:      "PAYMENT.VIRTUAL-ACCOUNT.PAID",
			Label:      "Payment Virtual Account Paid",
			EventGroup: "Payment",
			IsActive:   true,
		},
	}

	tests := []struct {
		name       string
		setupMock  func(db *mysqlMocks.IMySqlExt, log *pdkLogMock.ILogger)
		wantError  bool
		wantResult []callbackModel.CallbackEvent
	}{
		{
			name: "SUCCESS: Returns events ordered by group and event",
			setupMock: func(db *mysqlMocks.IMySqlExt, log *pdkLogMock.ILogger) {
				db.On("SelectContext", mock.Anything, mock.AnythingOfType(ArrCallbackEvent), mock.AnythingOfType("string")).
					Run(func(args mock.Arguments) {
						*args.Get(1).(*[]callbackModel.CallbackEvent) = expectedEvents
					}).
					Return(nil)
			},
			wantError:  false,
			wantResult: expectedEvents,
		},
		{
			name: "ERROR: Database returns error",
			setupMock: func(db *mysqlMocks.IMySqlExt, log *pdkLogMock.ILogger) {
				db.On("SelectContext", mock.Anything, mock.AnythingOfType(ArrCallbackEvent), mock.AnythingOfType("string")).
					Return(assert.AnError)
				log.On("Error", mock.Anything, "error when getting callback events", mock.Anything).Return()
			},
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := mysqlMocks.NewIMySqlExt(t)
			log := pdkLogMock.NewILogger(t)
			repo := New(db, log)

			test.setupMock(db, log)

			result, err := repo.GetCallbackEvents(t.Context())

			if test.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, test.wantResult, result)
			}

			db.AssertExpectations(t)
			log.AssertExpectations(t)
		})
	}
}
