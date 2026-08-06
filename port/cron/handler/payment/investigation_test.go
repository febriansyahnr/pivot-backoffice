package payment_test

import (
	"testing"
	"time"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/payment"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestProcessInvestigationMonthlyReconciliation(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	paymentSvc := serviceMocks.NewIPaymentService(t)

	handler := New(logger, paymentSvc)

	tests := []struct {
		name      string
		dateStr   string
		setupMock func()
	}{
		{
			name:    "ERROR:Invalid date format", // NOSONAR
			dateStr: "2026-02-05",
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Starting payment investigation monthly reconciliation",
				).Once().Return()
				logger.On(
					"Error", mock.Anything, "Invalid date format. Value: 2026-02-05 must be in YYYY-mm-dd HH:mm:ss format", mock.Anything,
				).Once().Return()
			},
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Starting payment investigation monthly reconciliation",
				).Once().Return()
				paymentSvc.On(
					"ProcessInvestigationMonthlyReconciliation", mock.Anything, mock.MatchedBy(func(req paymentModel.MonthlyReconciliationRequest) bool {
						return req.StartDate.Before(req.EndDate)
					}),
				).Once().Return(assert.AnError)
				logger.On(
					"Info", mock.Anything, "Payment investigation monthly reconciliation completed", mock.Anything, mock.Anything, mock.Anything, zap.Bool("completed", false), zap.Error(assert.AnError),
				).Once().Return()
			},
		},
		{
			name:    "SUCCESS", // NOSONAR
			dateStr: "2026-02-05 00:01:00",
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Starting payment investigation monthly reconciliation",
				).Once().Return()
				paymentSvc.On(
					"ProcessInvestigationMonthlyReconciliation", mock.Anything, mock.MatchedBy(func(req paymentModel.MonthlyReconciliationRequest) bool {
						return req.StartDate.Equal(time.Date(2026, 1, 4, 17, 0, 0, 0, time.UTC)) &&
							req.EndDate.Equal(time.Date(2026, 2, 4, 16, 59, 59, 0, time.UTC))
					}),
				).Once().Return(nil)
				logger.On(
					"Info", mock.Anything, "Payment investigation monthly reconciliation completed", mock.Anything, mock.Anything, mock.Anything, zap.Bool("completed", true), zap.Error(nil),
				).Once().Return()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			handler.ProcessInvestigationMonthlyReconciliation(t.Context(), test.dateStr)

			logger.AssertExpectations(t)
			paymentSvc.AssertExpectations(t)
		})
	}
}
