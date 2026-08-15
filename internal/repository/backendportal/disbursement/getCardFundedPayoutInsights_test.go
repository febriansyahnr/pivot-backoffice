package disbursementRepository_test

import (
	"testing"
	"time"

	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/cardFundedPayout"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/backendportal/disbursement"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetCardFundedPayoutInsights(t *testing.T) {
	logger := mocks.NewILogger(t)
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, logger)

	tests := []struct {
		name       string
		filter     *cardFundedPayoutModel.FilterGetPayoutInsights
		setupMock  func()
		wantErr    bool
		wantResult *cardFundedPayoutModel.GetPayoutInsightsDTO
	}{
		{
			name:   "ERROR: Database query fails",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-123"},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "failed to get card funded payout insights", mock.Anything).Once().Return()
			},
			wantErr: true,
		},
		{
			name:   "SUCCESS: No waiting payouts",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-123"},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*cardFundedPayoutModel.GetPayoutInsightsDTO) = cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 0,
						Sum:   0,
					}
				}).Return(nil)
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsDTO{
				Count: 0,
				Sum:   0,
			},
		},
		{
			name:   "SUCCESS: With waiting payouts",
			filter: &cardFundedPayoutModel.FilterGetPayoutInsights{MerchantID: "merchant-123"},
			setupMock: func() {
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*cardFundedPayoutModel.GetPayoutInsightsDTO) = cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 3,
						Sum:   5000000.00,
					}
				}).Return(nil)
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsDTO{
				Count: 3,
				Sum:   5000000.00,
			},
		},
		{
			name: "SUCCESS: With date filter",
			filter: func() *cardFundedPayoutModel.FilterGetPayoutInsights {
				start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
				end := time.Date(2026, 3, 31, 23, 59, 59, 0, time.UTC)
				return &cardFundedPayoutModel.FilterGetPayoutInsights{
					MerchantID:     "merchant-123",
					StartCreatedAt: &start,
					EndCreatedAt:   &end,
				}
			}(),
			setupMock: func() {
				// With date filter, query has 2 extra args (startDate, endDate)
				db.On("GetContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*cardFundedPayoutModel.GetPayoutInsightsDTO) = cardFundedPayoutModel.GetPayoutInsightsDTO{
						Count: 2,
						Sum:   3000000.00,
					}
				}).Return(nil)
			},
			wantErr: false,
			wantResult: &cardFundedPayoutModel.GetPayoutInsightsDTO{
				Count: 2,
				Sum:   3000000.00,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.GetCardFundedPayoutInsights(t.Context(), tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResult, result)
			}
		})
	}
}
