package recurringContractModel_test

import (
	"fmt"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestCreateRecurringContractRequestValidate(t *testing.T) {
	tests := []struct {
		name      string
		request   CreateRecurringContractRequest
		wantError error
	}{
		{
			name: "ERROR:Start of trial does not start from 1",
			request: CreateRecurringContractRequest{
				Trials: []Trial{
					{TrialStart: 2, TrialEnd: 2},
				},
			},
			wantError: fmt.Errorf("%s", "Ensure the initial trial value starts from 1"), // NOSONAR
		},
		{
			name: "ERROR:Non-consecutive trial periods",
			request: CreateRecurringContractRequest{
				Trials: []Trial{
					{TrialStart: 1, TrialEnd: 2},
					{TrialStart: 4, TrialEnd: 6},
				},
			},
			wantError: fmt.Errorf("%s", "Ensure the next trial value is sequential"), // NOSONAR
		},
		{
			name: "ERROR:Trial discount exceeds total transaction",
			request: CreateRecurringContractRequest{
				Trials: []Trial{
					{TrialStart: 1, TrialEnd: 3, Type: constant.RecurringContractTrialTypeFixed, Amount: 1},
				},
			},
			wantError: fmt.Errorf("%s", "Ensure the trial discount does not exceed the transaction amount"), // NOSONAR
		},
		{
			name: "ERROR:First authorization method is invalid",
			request: CreateRecurringContractRequest{
				Trials: []Trial{
					{TrialStart: 1, TrialEnd: 3, Type: constant.RecurringContractTrialTypeFree},
				},
				Amount: Amount{
					Value: 75_000.00, // NOSONAR
				},
				FirstAuthorization: constant.RecurringContractAuthMethodFirstPayment,
			},
			wantError: fmt.Errorf("%s", "First authorization must use the ONE_DOLLAR method because the first payment amount is 0"), // NOSONAR
		},
		{
			name: "ERROR:Can only use one of Customer Object or Customer ID",
			request: CreateRecurringContractRequest{
				Customer:   &unifiedPaymentModel.CustomerInformation{},
				CustomerID: util.ValueToPtr("f946a068-3c0c-4563-a331-894f9e3adef8"),
			},
			wantError: fmt.Errorf("%s", "Only one of Customer Object or Customer ID can be used"), // NOSONAR
		},
		{
			name: "SUCCESS",
			request: CreateRecurringContractRequest{
				Amount: Amount{
					Value: 75_000,
				},
				Trials: []Trial{
					{TrialStart: 1, TrialEnd: 1, Type: constant.RecurringContractTrialTypeFree},
					{TrialStart: 2, TrialEnd: 2, Type: constant.RecurringContractTrialTypeFixed, Amount: 5_000},
					{TrialStart: 3, TrialEnd: 3, Type: constant.RecurringContractTrialTypePercentage, Percentage: 10},
				},
				CustomerID:         util.ValueToPtr("e1706dd9-5239-4b3e-bfe9-ff53d126ba72"),
				FirstAuthorization: constant.RecurringContractAuthMethodOneDollar,
			},
			wantError: nil, // NOSONAR
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.wantError, test.request.Validate())
		})
	}
}
