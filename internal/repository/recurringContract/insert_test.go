package recurringContractRepo_test

import (
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/recurringContract"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInsert(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(nil, db)

	data := model.RecurringContract{
		Plan: model.Plan{
			PlanId:   "123456",
			PlanName: "Platinum",
		},
		Trials: []model.Trial{{
			TrialStart: 1,
			TrialEnd:   1,
			Type:       constant.RecurringContractTrialTypeFree,
		}},
		Billing: model.Billing{
			Interval:     1,
			IntervalUnit: constant.RecurringContractBillingIntervalUnitMonth,
		},
		Amount:  10_000,
		RawPlan: []byte(`{"planId":"123456","planName":"Platinum"}`),
		RawTrials: types.NullJSONText{
			Valid:    true,
			JSONText: []byte(`[{"trialStart":1,"trialEnd":1,"type":"FREE"}]`),
		},
		RawBilling: []byte(`{"interval":1,"intervalUnit":"MONTH","count":0}`),
	}

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Duplicate entry", // NOSONAR
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(false, fmt.Errorf("%s", "Error 1062 (23000): Duplicate entry 'aec6636d-7a02-4d93-a4c5-006b9c235068-RC/202601/000001' for key 'recurring_contracts.merchant_client_reference_comp_idx"))
			},
			wantError: constant.ErrClientReferenceIDAlreadyExist,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("NamedExecContext", mock.Anything, mock.Anything, data).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.Insert(t.Context(), data))
		})
	}
}
