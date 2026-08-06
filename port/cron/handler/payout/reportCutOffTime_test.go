package payout_test

import (
	"context"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/cron/handler/payout"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func TestReportAfterPayoutCutOffTime(t *testing.T) {
	log := loggerMock.NewILogger(t)
	disbursementSvc := serviceMock.NewIDisbursementService(t)

	report := disbursementModel.AfterPayoutCutOffTimeSummary{
		Total:  1,
		Amount: 10_000,
		Banks: []disbursementModel.AfterPayoutCutOffTimeBankSummary{
			{
				Name: "Bank Dummy", Total: 1, Amount: 10_000,
			},
		},
	}

	service := New(log, disbursementSvc)

	tests := []struct {
		name      string
		setupMock func()
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				log.On(
					"Info", c.ValueCtxMockType(), "Report after payout cut-off time",
					logger.String("status", "FAILED"),
					logger.Any("details", disbursementModel.AfterPayoutCutOffTimeSummary{}),
					logger.Error(c.ErrSomeErrorForUnitTest),
				).Once().Return()
				disbursementSvc.On(
					"ReportAfterPayoutCutOffTime", c.ValueCtxMockType(), c.TimeMockType(), c.TimeMockType(),
				).Once().Return(disbursementModel.AfterPayoutCutOffTimeSummary{}, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				log.On(
					"Info", c.ValueCtxMockType(), "Report after payout cut-off time",
					logger.String("status", "SUCCESS"), logger.Any("details", report), logger.Error(nil),
				).Once().Return()
				disbursementSvc.On(
					"ReportAfterPayoutCutOffTime", c.ValueCtxMockType(), c.TimeMockType(), c.TimeMockType(),
				).Return(report, nil)
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			service.ReportAfterPayoutCutOffTime(context.Background())

			log.AssertExpectations(t)
			disbursementSvc.AssertExpectations(t)
		})
	}
}
