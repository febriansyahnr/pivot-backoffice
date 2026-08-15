package reconciliation_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/reconciliation"
	"github.com/stretchr/testify/assert"
)

func TestReconciliationFilterRequestQuery(t *testing.T) {
	tests := []struct {
		name     string
		request  reconciliation.ReconciliationFilterRequest
		expected string
	}{
		{
			name:     "empty filter should return empty query",
			request:  reconciliation.ReconciliationFilterRequest{},
			expected: "",
		},
		{
			name: "filter with status should return status query",
			request: reconciliation.ReconciliationFilterRequest{
				Status: "PENDING",
			},
			expected: "WHERE status = 'PENDING'",
		},
		{
			name: "filter with empty status should return empty query",
			request: reconciliation.ReconciliationFilterRequest{
				Status: "",
			},
			expected: "",
		},
		{
			name: "filter with special characters in status should be preserved",
			request: reconciliation.ReconciliationFilterRequest{
				Status: "IN_PROGRESS",
			},
			expected: "WHERE status = 'IN_PROGRESS'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.Query()
			assert.Equal(t, tt.expected, result)
		})
	}
}
