package recurringContractModel_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/recurringContract"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestRecurringContractDetail(t *testing.T) {
	t.Run("First authorization status", func(t *testing.T) {
		tests := []struct {
			status     string
			wantResult bool
		}{
			{status: constant.RecurringContractStatusCreated, wantResult: true},
			{status: constant.RecurringContractStatusPendInitialAuth, wantResult: true},
			{status: constant.RecurringContractStatusActive, wantResult: false},
		}
		for _, test := range tests {
			rc := RecurringContractDetail{Status: test.status}
			assert.Equal(t, test.wantResult, rc.IsFirstAuthorization())
		}
	})

	t.Run("Get recurring amount for billing cycle", func(t *testing.T) {
		tests := []struct {
			initiate          bool
			recurringContract RecurringContractDetail
			wantAmount        float64
		}{
			{
				initiate: true,
				recurringContract: RecurringContractDetail{
					Status:     constant.RecurringContractStatusCreated,
					AuthMethod: constant.RecurringContractAuthMethodOneDollar,
				},
				wantAmount: 10_000.00, // NOSONAR
			},
			{
				initiate: true,
				recurringContract: RecurringContractDetail{
					Status:     constant.RecurringContractStatusPendInitialAuth,
					AuthMethod: constant.RecurringContractAuthMethodFirstPayment,
					Amount:     75_000.00, // NOSONAR
				},
				wantAmount: 75_000.00, // NOSONAR
			},
			{
				initiate: false,
				recurringContract: RecurringContractDetail{
					Status: constant.RecurringContractStatusActive,
					Amount: 125_000.00, // NOSONAR
				},
				wantAmount: 125_000.00, // NOSONAR
			},
			{
				initiate: false,
				recurringContract: RecurringContractDetail{
					Status: constant.RecurringContractStatusActive,
					Trials: []Trial{
						{
							TrialStart: 1,
							TrialEnd:   3,
							Type:       constant.RecurringContractTrialTypeFree,
						},
					},
					Billing: Billing{
						Count: 0,
					},
					Amount: 125_000.00, // NOSONAR
				},
				wantAmount: 0.00, // NOSONAR
			},
			{
				initiate: false,
				recurringContract: RecurringContractDetail{
					Status: constant.RecurringContractStatusActive,
					Trials: []Trial{
						{
							TrialStart: 1,
							TrialEnd:   3,
							Type:       constant.RecurringContractTrialTypeFixed,
							Amount:     10_000,
						},
					},
					Billing: Billing{
						Count: 1,
					},
					Amount: 125_000.00, // NOSONAR
				},
				wantAmount: 115_000.00, // NOSONAR
			},
			{
				initiate: false,
				recurringContract: RecurringContractDetail{
					Status: constant.RecurringContractStatusActive,
					Trials: []Trial{
						{
							TrialStart: 1,
							TrialEnd:   3,
							Type:       constant.RecurringContractTrialTypePercentage,
							Percentage: 25,
						},
					},
					Billing: Billing{
						Count: 2,
					},
					Amount: 100_000.00, // NOSONAR
				},
				wantAmount: 75_000.00, // NOSONAR
			},
		}
		for _, test := range tests {
			assert.Equal(t, test.wantAmount, test.recurringContract.GetRecurringAmountForBillingCycle(test.initiate))
		}
	})

	t.Run("Get unified payment method type", func(t *testing.T) {
		tests := []struct {
			paymentMethodType *string
			wantResult        string
		}{
			{paymentMethodType: nil},
			{paymentMethodType: util.ValueToPtr("CREDIT_CARD"), wantResult: constant.UnifiedPaymentMethodCard},
			{paymentMethodType: util.ValueToPtr(constant.UnifiedPaymentMethodEWallet), wantResult: constant.UnifiedPaymentMethodEWallet},
		}
		for _, test := range tests {
			rc := RecurringContractDetail{
				PaymentMethodType: test.paymentMethodType,
			}
			assert.Equal(t, test.wantResult, rc.GetUnifiedPaymentMethodType())
		}
	})

	t.Run("Get min max amount per payment", func(t *testing.T) {
		tests := []struct {
			amount        float64
			trials        []Trial
			wantMinAmount float64
			wantMaxAmount float64
		}{
			{
				amount:        75_000, // NOSONAR
				wantMinAmount: 75_000, // NOSONAR
				wantMaxAmount: 75_000, // NOSONAR
			},
			{
				amount: 125_000, // NOSONAR
				trials: []Trial{
					{
						TrialStart: 1,
						TrialEnd:   3,
						Type:       constant.RecurringContractTrialTypeFree,
					},
				},
				wantMinAmount: 0,       // NOSONAR
				wantMaxAmount: 125_000, // NOSONAR
			},
			{
				amount: 100_000, // NOSONAR
				trials: []Trial{
					{
						TrialStart: 1,
						TrialEnd:   1,
						Type:       constant.RecurringContractTrialTypePercentage,
						Percentage: 50,
					},
					{
						TrialStart: 1,
						TrialEnd:   1,
						Type:       constant.RecurringContractTrialTypePercentage,
						Percentage: 90,
					},
					{
						TrialStart: 1,
						TrialEnd:   1,
						Type:       constant.RecurringContractTrialTypePercentage,
						Percentage: 75,
					},
				},
				wantMinAmount: 10_000,  // NOSONAR
				wantMaxAmount: 100_000, // NOSONAR
			},
		}
		for _, test := range tests {
			rc := RecurringContractDetail{
				Amount: test.amount,
				Trials: test.trials,
			}
			minAmount, maxAmount := rc.GetMinMaxAmountPerPayment()

			assert.Equal(t, test.wantMinAmount, minAmount)
			assert.Equal(t, test.wantMaxAmount, maxAmount)
		}
	})
}
