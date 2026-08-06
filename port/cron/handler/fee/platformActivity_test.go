package feeHandler_test

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/fee"

	"github.com/stretchr/testify/assert"
)

func TestPlatformActivitiesFee(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIFeeService(t)

	handler := New(logger, service)

	tests := []struct {
		name      string
		date      string
		setupMock func()
	}{
		{
			name: "ERROR:Datetime format",
			date: "2024-09-09",
			setupMock: func() {
				logger.On(
					"Error", c.ValueCtxMockType(), "Invalid date format. Value: 2024-09-09 must be in YYYY-mm-dd HH:MM:SS format", c.ZapFieldMockType(),
				).Once().Return()
			},
		},
		{
			name: "ERROR:Some error",
			date: "2024-09-09 00:15:02",
			setupMock: func() {
				logger.On(
					"Fatal", c.ValueCtxMockType(), "An error occurred while calculating platform activity fees", c.ZapFieldMockType(),
				).Once().Return()
				service.On(
					"PlatformActivitiesFee", c.ValueCtxMockType(), time.Date(2024, 9, 9, 0, 15, 2, 0, tz),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS:Custom date",
			date: "2024-08-28 03:46:12",
			setupMock: func() {
				service.On("PlatformActivitiesFee", c.ValueCtxMockType(), time.Date(2024, 8, 28, 3, 46, 12, 0, tz)).Once().Return(nil)
			},
		},
		{
			name: "SUCCESS:Current date",
			setupMock: func() {
				service.On("PlatformActivitiesFee", c.ValueCtxMockType(), c.TimeMockType()).Once().Return(nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger.On(
				"Info", c.ValueCtxMockType(), "Start the fee calculation process for platform activities",
			).Times(1).Return()
			logger.On(
				"Info", c.ValueCtxMockType(), "Platform activity fee calculation completed", c.ZapFieldMockType(), c.ZapFieldMockType(), c.ZapFieldMockType(),
			).Times(1).Return()

			test.setupMock()
			handler.PlatformActivitiesFee(context.Background(), test.date)

			assert.True(t, logger.AssertExpectations(t))
			assert.True(t, service.AssertExpectations(t))
		})
	}
}
