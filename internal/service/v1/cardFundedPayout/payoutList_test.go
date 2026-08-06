package cardFundedPayoutService

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	cardFundedPayoutModel "github.com/paper-indonesia/pivot-backoffice/internal/model/cardFundedPayout"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPayoutList(t *testing.T) {
	cfg := &config.Config{}
	disbursementRepo := repositoryMock.NewIDisbursementRepository(t)
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})

	svc := New(cfg, log,
		WithDisbursementRepository(disbursementRepo),
	)

	validFilter := &cardFundedPayoutModel.FilterGetPayoutList{
		MerchantID: "merchant-123",
	}

	testCases := []struct {
		name      string
		filter    *cardFundedPayoutModel.FilterGetPayoutList
		setupMock func()
		wantErr   bool
	}{
		{
			name:   "ERROR: Get card funded payout list failed",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(nil, errors.New("database error")).Once()
			},
			wantErr: true,
		},
		{
			name:   "SUCCESS: Get payout list with empty result",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*cardFundedPayoutModel.GetPayoutListResponse{},
						Meta: commonModel.Meta{
							TotalItems: 0,
						},
					}, nil).Once()
			},
			wantErr: false,
		},
		{
			name:   "SUCCESS: Get payout list with data",
			filter: validFilter,
			setupMock: func() {
				disbursementRepo.On("GetCardFundedPayoutList", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{
						Data: []*cardFundedPayoutModel.GetPayoutListResponse{
							{
								UUID:        "payout-123",
								ReferenceID: "ref-123",
							},
						},
						Meta: commonModel.Meta{
							TotalItems: 1,
						},
					}, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := svc.GetPayoutList(context.Background(), tc.filter)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			disbursementRepo.AssertExpectations(t)
		})
	}
}
