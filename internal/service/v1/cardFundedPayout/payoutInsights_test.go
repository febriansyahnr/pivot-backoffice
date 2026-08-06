package cardFundedPayoutService

import (
	"context"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutInsights(t *testing.T) {
	cfg := &config.Config{}
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
	)

	testCases := []struct {
		name       string
		filter     *cardFundedPayoutModel.FilterGetPayoutInsights
		setupMock  func()
		wantErr    bool
		wantResult *cardFundedPayoutModel.GetPayoutInsightsResponse
	}{
		{
			name:   "ERROR: Repository returns error",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-123"},
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutInsights", mock.Anything, mock.Anything).
					Return(nil, assert.AnError).Once()
			},
			wantErr: true,
		},
		{
			name:   "SUCCESS: No waiting payouts",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-123"},
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutInsights", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 0,
						Sum:   0,
					}, nil).Once()
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsResponse{
				TotalTransaction: 0,
				TotalAmount: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "0.00",
				},
			},
		},
		{
			name:   "SUCCESS: With waiting payouts",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-456"},
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutInsights", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 5,
						Sum:   15000000.50,
					}, nil).Once()
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsResponse{
				TotalTransaction: 5,
				TotalAmount: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "15000000.50",
				},
			},
		},
		{
			name: "SUCCESS: With date filter",
			filter: func() *cardFundedPayoutModel.FilterGetPayoutInsights {
				start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
				return &cardFundedPayoutModel.FilterGetPayoutInsights{
					MerchantID:     "merchant-789",
					StartCreatedAt: &start,
					EndCreatedAt:   &end,
				}
			}(),
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutInsights", mock.Anything, mock.Anything).
					Return(&cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 2,
						Sum:   3000000.00,
					}, nil).Once()
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsResponse{
				TotalTransaction: 2,
				TotalAmount: commonModel.Amount{
					Currency: constant.CurrencyIDR,
					Value:    "3000000.00",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.GetPayoutInsights(context.Background(), tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			}

			disbursementRepo.AssertExpectations(t)
		})
	}
}
