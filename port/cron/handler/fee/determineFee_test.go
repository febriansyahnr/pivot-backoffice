package feeHandler_test

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/fee"

	"go.uber.org/zap"
)

func TestDetermineFeeTierLvlFromMonthlyTPV(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	feeService := serviceMocks.NewIFeeService(t)

	handler := New(logger, feeService)

	nowInTz := time.Now().In(tz)

	tests := []struct {
		name      string
		date      string
		setupMock func()
	}{
		{
			name: "ERROR:Invalid date input",
			date: "2024-10",
			setupMock: func() {
				logger.On(
					"Error", c.ValueCtxMockType(), "Invalid date format. Value: 2024-10 must be in YYYY-mm-dd format", c.ZapFieldMockType(),
				).Once().Return()
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				logger.On(
					"Info", c.ValueCtxMockType(), "Start determine fee tier level from monthly TPV period "+nowInTz.Format("2006-01"),
				).Times(1).Return()
				feeService.On(
					"DetermineFeeTierLvlFromMonthlyTPV", c.ValueCtxMockType(), c.TimeMockType(),
				).Once().Return(c.ErrSomeErrorForUnitTest)
				logger.On(
					"Fatal", c.ValueCtxMockType(), "An error occurred while determine fee tier level from monthly tpv", c.ZapFieldMockType(),
				).Once().Return()
				logger.On(
					"Info", c.ValueCtxMockType(), "Determine fee tier level from monthly TPV completed", c.ZapFieldMockType(), c.ZapFieldMockType(), zap.Bool("completed", false),
				).Times(1).Return()
			},
		},
		{
			name: "SUCCESS", // NOSONAR
			date: "2024-10-01",
			setupMock: func() {
				logger.On(
					"Info", c.ValueCtxMockType(), "Start determine fee tier level from monthly TPV period 2024-10",
				).Times(1).Return()
				feeService.On("DetermineFeeTierLvlFromMonthlyTPV", c.ValueCtxMockType(), c.TimeMockType()).Return(nil)
				logger.On(
					"Info", c.ValueCtxMockType(), "Determine fee tier level from monthly TPV completed", c.ZapFieldMockType(), c.ZapFieldMockType(), zap.Bool("completed", true),
				).Times(1).Return()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			handler.DetermineFeeTierLvlFromMonthlyTPV(context.Background(), test.date)

			logger.AssertExpectations(t)
			feeService.AssertExpectations(t)
		})
	}
}
