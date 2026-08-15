package reconciliation

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestGetFirstLatestDate(t *testing.T) {
	tests := []struct {
		name           string
		transactions   []*Transaction
		wantFirstDate  *time.Time
		wantLatestDate *time.Time
	}{
		{
			name:           "empty transactions",
			transactions:   []*Transaction{},
			wantFirstDate:  nil,
			wantLatestDate: nil,
		},
		{
			name: "single transaction",
			transactions: []*Transaction{
				{TransactionDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			},
			wantFirstDate:  timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			wantLatestDate: timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "multiple transactions in order",
			transactions: []*Transaction{
				{TransactionDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{TransactionDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
				{TransactionDate: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)},
			},
			wantFirstDate:  timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			wantLatestDate: timePtr(time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
		{
			name: "multiple transactions out of order",
			transactions: []*Transaction{
				{TransactionDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)},
				{TransactionDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
				{TransactionDate: time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)},
			},
			wantFirstDate:  timePtr(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			wantLatestDate: timePtr(time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconResult{
				Transactions: tt.transactions,
			}
			gotFirstDate, gotLatestDate := r.GetFirstLatestDate()

			if tt.wantFirstDate == nil {
				require.Nil(t, gotFirstDate)
			} else {
				require.Equal(t, *tt.wantFirstDate, *gotFirstDate)
			}

			if tt.wantLatestDate == nil {
				require.Nil(t, gotLatestDate)
			} else {
				require.Equal(t, *tt.wantLatestDate, *gotLatestDate)
			}
		})
	}
}

func TestShouldReconcileVAStatic(t *testing.T) {
	tests := []struct {
		name     string
		vaStatic *ReconVAStatic
		want     bool
	}{
		{
			name:     "nil VAStatic",
			vaStatic: nil,
			want:     false,
		},
		{
			name:     "empty VAStatic",
			vaStatic: &ReconVAStatic{},
			want:     false,
		},
		{
			name: "VAStatic with data",
			vaStatic: &ReconVAStatic{
				"reference1": &ReconVAStaticResult{
					Indexes: []int{0},
					UUIDs:   []string{"uuid1"},
					IsValid: true,
				},
			},
			want: true,
		},
		{
			name: "VAStatic with multiple entries",
			vaStatic: &ReconVAStatic{
				"reference1": &ReconVAStaticResult{
					Indexes: []int{0},
					UUIDs:   []string{"uuid1"},
					IsValid: true,
				},
				"reference2": &ReconVAStaticResult{
					Indexes: []int{1},
					UUIDs:   []string{"uuid2"},
					IsValid: false,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ReconResult{
				VAStatic: tt.vaStatic,
			}
			got := r.ShouldReconcileVAStatic()
			require.Equal(t, tt.want, got)
		})
	}
}

func TestReconVAStatic_Add(t *testing.T) {
	tests := []struct {
		name              string
		initialVAStatic   ReconVAStatic
		reference         string
		index             int
		amount            decimal.Decimal
		trx               *ReconTransactionModel
		expectedVAStatic  ReconVAStatic
	}{
		{
			name:            "add new reference to empty VAStatic",
			initialVAStatic: ReconVAStatic{},
			reference:       "ref1",
			index:           0,
			amount:          decimal.NewFromInt(100),
			trx:             &ReconTransactionModel{UUID: "uuid1"},
			expectedVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
		},
		{
			name: "add to existing reference",
			initialVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			reference: "ref1",
			index:     1,
			amount:    decimal.NewFromInt(50),
			trx:       &ReconTransactionModel{UUID: "uuid2"},
			expectedVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0, 1},
					UUIDs:       []string{"uuid1", "uuid2"},
					TotalAmount: decimal.NewFromInt(150),
				},
			},
		},
		{
			name: "add new reference to existing VAStatic with data",
			initialVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			reference: "ref2",
			index:     1,
			amount:    decimal.NewFromInt(200),
			trx:       &ReconTransactionModel{UUID: "uuid2"},
			expectedVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
				"ref2": &ReconVAStaticResult{
					Indexes:     []int{1},
					UUIDs:       []string{"uuid2"},
					TotalAmount: decimal.NewFromInt(200),
				},
			},
		},
		{
			name: "add multiple entries to same reference",
			initialVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			reference: "ref1",
			index:     2,
			amount:    decimal.NewFromInt(75),
			trx:       &ReconTransactionModel{UUID: "uuid3"},
			expectedVAStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0, 2},
					UUIDs:       []string{"uuid1", "uuid3"},
					TotalAmount: decimal.NewFromInt(175),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaStatic := tt.initialVAStatic
			vaStatic.Add(tt.reference, tt.index, tt.amount, tt.trx)

			require.Equal(t, len(tt.expectedVAStatic), len(vaStatic))
			
			for ref, expected := range tt.expectedVAStatic {
				actual, exists := vaStatic[ref]
				require.True(t, exists, "reference %s should exist", ref)
				require.Equal(t, expected.Indexes, actual.Indexes)
				require.Equal(t, expected.UUIDs, actual.UUIDs)
				require.True(t, expected.TotalAmount.Equal(actual.TotalAmount))
			}
		})
	}
}

func TestReconVAStatic_Keys(t *testing.T) {
	tests := []struct {
		name        string
		vaStatic    ReconVAStatic
		expectedLen int
		contains    []string
	}{
		{
			name:        "empty VAStatic",
			vaStatic:    ReconVAStatic{},
			expectedLen: 0,
			contains:    []string{},
		},
		{
			name: "single key",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			expectedLen: 1,
			contains:    []string{"ref1"},
		},
		{
			name: "multiple keys",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
				"ref2": &ReconVAStaticResult{
					Indexes:     []int{1},
					UUIDs:       []string{"uuid2"},
					TotalAmount: decimal.NewFromInt(200),
				},
				"ref3": &ReconVAStaticResult{
					Indexes:     []int{2},
					UUIDs:       []string{"uuid3"},
					TotalAmount: decimal.NewFromInt(300),
				},
			},
			expectedLen: 3,
			contains:    []string{"ref1", "ref2", "ref3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := tt.vaStatic.Keys()
			require.Equal(t, tt.expectedLen, len(keys))
			
			for _, expectedKey := range tt.contains {
				require.Contains(t, keys, expectedKey)
			}
		})
	}
}

func TestReconVAStatic_GetIndexes(t *testing.T) {
	tests := []struct {
		name            string
		vaStatic        ReconVAStatic
		reference       string
		expectedIndexes []int
	}{
		{
			name:            "empty VAStatic",
			vaStatic:        ReconVAStatic{},
			reference:       "ref1",
			expectedIndexes: []int{},
		},
		{
			name: "reference not found",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0, 1},
					UUIDs:       []string{"uuid1", "uuid2"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			reference:       "ref2",
			expectedIndexes: []int{},
		},
		{
			name: "reference found with single index",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0},
					UUIDs:       []string{"uuid1"},
					TotalAmount: decimal.NewFromInt(100),
				},
			},
			reference:       "ref1",
			expectedIndexes: []int{0},
		},
		{
			name: "reference found with multiple indexes",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0, 2, 5},
					UUIDs:       []string{"uuid1", "uuid2", "uuid3"},
					TotalAmount: decimal.NewFromInt(300),
				},
			},
			reference:       "ref1",
			expectedIndexes: []int{0, 2, 5},
		},
		{
			name: "reference found among multiple references",
			vaStatic: ReconVAStatic{
				"ref1": &ReconVAStaticResult{
					Indexes:     []int{0, 1},
					UUIDs:       []string{"uuid1", "uuid2"},
					TotalAmount: decimal.NewFromInt(100),
				},
				"ref2": &ReconVAStaticResult{
					Indexes:     []int{3, 4, 6},
					UUIDs:       []string{"uuid3", "uuid4", "uuid5"},
					TotalAmount: decimal.NewFromInt(200),
				},
			},
			reference:       "ref2",
			expectedIndexes: []int{3, 4, 6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexes := tt.vaStatic.GetIndexes(tt.reference)
			require.Equal(t, tt.expectedIndexes, indexes)
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
