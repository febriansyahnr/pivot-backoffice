package callbackService_test

import (
	"testing"

	callbackModel "github.com/paper-indonesia/pivot-backoffice/internal/model/callback"
	callbackService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/callback"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetCallbackEvents(t *testing.T) {
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
			Event:      "PAYMENT.VIRTUAL-ACCOUNT.PAID",
			Label:      "Payment Virtual Account Paid",
			EventGroup: "Payment",
			IsActive:   true,
		},
		{
			UUID:       uuid.MustParse("550e8400-e29b-41d4-a716-446655440002"),
			Event:      "REFUND.SUCCESS",
			Label:      "Refund Success",
			EventGroup: "Refund",
			IsActive:   true,
		},
	}

	testCases := []struct {
		name           string
		wantErr        bool
		expectedResult []callbackModel.CallbackEvent
		expectedError  error
		mockSetup      func(mockRepo *repositoryMocks.ICallbackRepository, mockLogger *pdkLogMock.ILogger)
	}{
		{
			name:           "SUCCESS: Get all callback events",
			wantErr:        false,
			expectedResult: expectedEvents,
			mockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, mockLogger *pdkLogMock.ILogger) {
				mockRepo.On(
					"GetCallbackEvents",
					mock.Anything,
				).Return(expectedEvents, nil)
			},
		},
		{
			name:           "SUCCESS: Get empty callback events list",
			wantErr:        false,
			expectedResult: []callbackModel.CallbackEvent{},
			mockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, mockLogger *pdkLogMock.ILogger) {
				mockRepo.On(
					"GetCallbackEvents",
					mock.Anything,
				).Return([]callbackModel.CallbackEvent{}, nil)
			},
		},
		{
			name:          "ERROR: Repository returns error",
			wantErr:       true,
			expectedError: assert.AnError,
			mockSetup: func(mockRepo *repositoryMocks.ICallbackRepository, mockLogger *pdkLogMock.ILogger) {
				mockRepo.On(
					"GetCallbackEvents",
					mock.Anything,
				).Return(nil, assert.AnError)
				mockLogger.On(
					"Error",
					mock.Anything,
					"failed to get callback events",
					mock.Anything,
				).Return()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := repositoryMocks.NewICallbackRepository(t)
			mockLogger := pdkLogMock.NewILogger(t)
			tc.mockSetup(mockRepo, mockLogger)

			cbService := callbackService.New(mockLogger, nil, mockRepo, nil, nil)
			result, err := cbService.GetCallbackEvents(t.Context())

			if tc.wantErr {
				require.Nil(t, result)
				require.Error(t, err)
				require.Equal(t, tc.expectedError, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, tc.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	}
}
