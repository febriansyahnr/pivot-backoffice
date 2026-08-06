package paymentService_test

import (
	"testing"
	"time"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentDashboardInsights(t *testing.T) {
	paymentRepo := repoMocks.NewIPaymentRepository(t)
	paymentMetrics := repoMocks.NewIDatamartPaymentMetrics(t)

	service := New(paymentRepo, nil, nil, nil, nil, nil, nil, WithPaymentMetricsRepository(paymentMetrics))

	request := paymentModel.GetPaymentDashboardInsightRequest{
		MerchantId: "d39896de-2169-4fca-979a-8d239af7ab30",
		StartDate:  time.Date(2025, 11, 23, 17, 00, 00, 0, time.UTC),
		EndDate:    time.Date(2025, 11, 30, 16, 59, 59, 999, time.UTC),
	}
	successRateRequest := paymentModel.QueryPaymentSuccessRateComparisonRequest{
		MerchantId: request.MerchantId,
		PrevRange: paymentModel.DateRange{
			StartDate: "2025-11-17",
			EndDate:   "2025-11-23",
		},
		CurrentRange: paymentModel.DateRange{
			StartDate: "2025-11-24",
			EndDate:   "2025-11-30",
		},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *paymentModel.PaymentDashboardInsights
	}{
		{
			name: "ERROR:Aggregate payment", // NOSONAR
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentDashboardInsights", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
				paymentMetrics.On(
					"QueryPaymentSuccessRateComparison", mock.Anything, successRateRequest,
				).Once().Return(nil, nil)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Query success rate", // NOSONAR
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentDashboardInsights", mock.Anything, mock.Anything,
				).Once().Return(&paymentModel.PaymentDashboardInsights{}, nil)
				paymentMetrics.On(
					"QueryPaymentSuccessRateComparison", mock.Anything, successRateRequest,
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				paymentRepo.On(
					"GetPaymentDashboardInsights", mock.Anything, mock.Anything,
				).Once().Return(&paymentModel.PaymentDashboardInsights{}, nil)
				paymentMetrics.On(
					"QueryPaymentSuccessRateComparison", mock.Anything, successRateRequest,
				).Once().Return(&paymentModel.PaymentSuccessRateComparison{}, nil)
			},
			wantResult: &paymentModel.PaymentDashboardInsights{
				SuccessRate: &paymentModel.PaymentSuccessRateComparison{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetPaymentDashboardInsights(t.Context(), request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
