package constant_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/constant"

	"github.com/stretchr/testify/assert"
)

func TestRecurringMinDaysBetweenPayments(t *testing.T) {
	tests := []struct {
		intervalUnit               string
		wantMinDaysBetweenPayments uint16
	}{
		{
			intervalUnit:               RecurringContractBillingIntervalUnitDay,
			wantMinDaysBetweenPayments: 1,
		},
		{
			intervalUnit:               RecurringContractBillingIntervalUnitMonth,
			wantMinDaysBetweenPayments: 28,
		},
		{
			intervalUnit:               RecurringContractBillingIntervalUnitYear,
			wantMinDaysBetweenPayments: 365,
		},
		{
			intervalUnit:               "OTHER",
			wantMinDaysBetweenPayments: 1,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantMinDaysBetweenPayments, RecurringMinDaysBetweenPayments(test.intervalUnit))
	}
}
