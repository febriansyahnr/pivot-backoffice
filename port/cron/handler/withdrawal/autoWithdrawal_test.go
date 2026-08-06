package withdrawalHandler_test

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/withdrawal"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestTriggeringAutoWithdrawalProcess(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(logger, service)

	tests := []struct {
		name      string
		messages  int64
		setupMock func()
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				service.On("TriggeringAutoWithdrawalProcess", c.ValueCtxMockType()).Once().Return(int64(0), c.ErrSomeErrorForUnitTest)
				logger.On(
					"Fatal", c.ValueCtxMockType(), "Failed when triggering auto withdrawal process", zap.Error(c.ErrSomeErrorForUnitTest),
				).Once().Return()
			},
		},
		{
			name:     "SUCCESS", // NOSONAR
			messages: 2,
			setupMock: func() {
				service.On("TriggeringAutoWithdrawalProcess", c.ValueCtxMockType()).Return(int64(2), nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			logger.On(
				"Info", c.ValueCtxMockType(), "Running the CronJob to trigger the automatic withdrawal process",
			).Times(1).Return()
			logger.On(
				"Info", c.ValueCtxMockType(), "CronJob completed successfully", zap.Int64("totalMessagesPublished", test.messages), c.ZapFieldMockType(),
			).Times(1).Return()

			handler.TriggeringAutoWithdrawalProcess(context.Background())

			logger.AssertExpectations(t)
			service.AssertExpectations(t)
		})
	}
}

func TestForceAutoWithdrawalProcess(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(logger, service)

	tests := []struct {
		name      string
		date      string
		setupMock func()
	}{
		{
			name: "ERROR:Date format",
			date: "2024-12-01",
			setupMock: func() {
				logger.On(
					"Error", c.ValueCtxMockType(), "Invalid date format. Value: 2024-12-01 must be in YYYY-mm-dd HH:mm:ss format", c.ZapFieldMockType(),
				).Once().Return()
			},
		},
		{
			name: "ERROR:Some error",
			date: "2024-12-03 00:15:32",
			setupMock: func() {
				logger.On(
					"Fatal", c.ValueCtxMockType(), "Failed when force auto withdrawal process", c.ZapFieldMockType(),
				).Once().Return()
				service.On(
					"ForceAutoWithdrawalProcess", c.ValueCtxMockType(), time.Date(2024, 12, 2, 17, 15, 32, 0, time.UTC),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS:Custom date",
			date: "2024-12-03 01:45:15",
			setupMock: func() {
				service.On(
					"ForceAutoWithdrawalProcess", c.ValueCtxMockType(), time.Date(2024, 12, 2, 18, 45, 15, 0, time.UTC),
				).Once().Return(nil, nil)
			},
		},
		{
			name: "SUCCESS:Current date",
			setupMock: func() {
				service.On(
					"ForceAutoWithdrawalProcess", c.ValueCtxMockType(), c.TimeMockType(),
				).Once().Return(nil, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger.On(
				"Info", c.ValueCtxMockType(), "Starting forced withdrawal process for dormant merchants",
			).Times(1).Return()
			logger.On(
				"Info", c.ValueCtxMockType(), "Forced withdrawal process for dormant merchants completed", c.ZapFieldMockType(), c.ZapFieldMockType(), c.ZapFieldMockType(), c.ZapFieldMockType(),
			).Times(1).Return()

			test.setupMock()
			handler.ForceAutoWithdrawalProcess(context.Background(), test.date)

			assert.True(t, logger.AssertExpectations(t))
			assert.True(t, service.AssertExpectations(t))
		})
	}
}
