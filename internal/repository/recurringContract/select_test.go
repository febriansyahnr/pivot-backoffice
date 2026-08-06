package recurringContractRepo_test

import (
	"database/sql"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/recurringContract"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetDetailByID(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(nil, db)

	merchantID := "e8549436-cb9c-4dd1-8813-5ab7147381e3"
	recurringID := "98ec1077-f20b-4a09-952d-4db9a76a2b00"

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *model.RecurringContractDetail
	}{
		{
			name: "ERROR:Database connection", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, recurringID,
				).Once().Return(sql.ErrConnDone)
			},
			wantError: sql.ErrConnDone,
		},
		{
			name: "ERROR:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, recurringID,
				).Once().Return(sql.ErrNoRows)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, recurringID,
				).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*model.RecurringContractDetail) = model.RecurringContractDetail{
						UUID:       recurringID,
						MerchantID: merchantID,
						RawTrials: types.NullJSONText{
							Valid: true, JSONText: []byte(`[{"type": "FREE", "trialEnd": 3, "trialStart": 1}]`),
						},
						RawBilling: []byte(`{"count": 2, "interval": 1, "intervalUnit": "MONTH"}`),
					}
				}).Return(nil)
			},
			wantResult: &model.RecurringContractDetail{
				UUID:       recurringID,
				MerchantID: merchantID,
				RawTrials: types.NullJSONText{
					Valid: true, JSONText: []byte(`[{"type": "FREE", "trialEnd": 3, "trialStart": 1}]`),
				},
				Trials: []model.Trial{
					{
						TrialStart: 1,
						TrialEnd:   3,
						Type:       constant.RecurringContractTrialTypeFree,
					},
				},
				RawBilling: []byte(`{"count": 2, "interval": 1, "intervalUnit": "MONTH"}`),
				Billing: model.Billing{
					Interval:     1,
					IntervalUnit: constant.RecurringContractBillingIntervalUnitMonth,
					Count:        2,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetDetailByID(t.Context(), merchantID, recurringID)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}
