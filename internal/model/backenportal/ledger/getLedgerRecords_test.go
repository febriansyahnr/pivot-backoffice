package ledger_model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdjustDateTime(t *testing.T) {

	testCases := []struct {
		Name      string
		StartDate time.Time
		EndDate   time.Time
		WantErr   bool
	}{
		{
			Name:      "SUCCESS: Valid Date Range",
			StartDate: time.Now().UTC(),
			EndDate:   time.Now().UTC().AddDate(0, 0, 7),
			WantErr:   false,
		},
		{
			Name:      "SUCCESS: Date range not defined",
			StartDate: time.Time{},
			EndDate:   time.Time{},
			WantErr:   false,
		},
		{
			Name:      "SUCCESS: End date not defined",
			StartDate: time.Now().UTC().AddDate(0, -1, 0),
			EndDate:   time.Time{},
			WantErr:   false,
		},
		{
			Name:      "SUCCESS: Start date not defined",
			EndDate:   time.Now().UTC().AddDate(0, -1, 0),
			StartDate: time.Time{},
			WantErr:   false,
		},
		{
			Name:      "ERROR: Start date bigger than end date",
			EndDate:   time.Now().UTC().AddDate(0, -1, 0),
			StartDate: time.Now(),
			WantErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			req := GetLedgerTransactionRequest{
				StartDate: tc.StartDate,
				EndDate:   tc.EndDate,
			}
			err := req.AdjustDateTime()
			if tc.WantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, time.Time{}, req.StartDate)
				assert.NotEqual(t, time.Time{}, req.EndDate)
			}
		})
	}
}
