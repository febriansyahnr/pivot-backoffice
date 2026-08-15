package disbursementRepository_test

import (
	"errors"
	"testing"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/disbursement"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetXbPayoutDashboardInsights(t *testing.T) {
	logger := mocks.NewILogger(t)
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, logger)
	args := []any{
		mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.XbPayoutDashboardInsights
	}{
		{
			name: "ERROR:Executing aggregate query", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "failed while executing aggregate query", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Unmarshal json", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*model.XbPayoutDashboardInsights) = model.XbPayoutDashboardInsights{
						RawTopCountriesByVolume: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`B`),
						},
					}
				}).Return(nil)
				logger.On("Error", mock.Anything, "failed while unmarshal json top countries by volume", mock.Anything).Once().Return()
			},
			wantError: errors.New("unmarshal json: invalid character 'B' looking for beginning of value"), // NOSONAR
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*model.XbPayoutDashboardInsights) = model.XbPayoutDashboardInsights{
						SuccessCount: 1,          // NOSONAR
						SuccessTotal: "10000.00", // NOSONAR
						RawTopCountriesByVolume: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(`[{"volume": 10000.00, "country": "Malaysia", "percentage": 100.00}, {"volume": 0.00, "currency": "OTHERS", "percentage": 0.00}]`),
						},
					}
				}).Return(nil)
			},
			wantResult: &model.XbPayoutDashboardInsights{
				SuccessCount: 1,          // NOSONAR
				SuccessTotal: "10000.00", // NOSONAR
				RawTopCountriesByVolume: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(`[{"volume": 10000.00, "country": "Malaysia", "percentage": 100.00}, {"volume": 0.00, "currency": "OTHERS", "percentage": 0.00}]`),
				},
				TopCountriesByVolume: []model.XbPayoutTransactionVolumeByCountry{
					{
						Country:    "Malaysia",
						Volume:     "10000.00",
						Percentage: "100.00",
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetXbPayoutDashboardInsights(t.Context(), model.GetXbPayoutDashboardInsightRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
