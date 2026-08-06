package xbPayoutService_test

import (
	"testing"

	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/xbPayout"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetXbPayoutDashboardInsights(t *testing.T) {

	payoutRepo := repoMocks.NewIDisbursementRepository(t)

	service := New(nil, payoutRepo, nil, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *disbursementModel.XbPayoutDashboardInsights
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				payoutRepo.On(
					"GetXbPayoutDashboardInsights", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				payoutRepo.On(
					"GetXbPayoutDashboardInsights", mock.Anything, mock.Anything,
				).Once().Return(&disbursementModel.XbPayoutDashboardInsights{}, nil)
			},
			wantResult: &disbursementModel.XbPayoutDashboardInsights{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetXbPayoutDashboardInsights(t.Context(), disbursementModel.GetXbPayoutDashboardInsightRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
