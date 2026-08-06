package commonModel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSortOrder(t *testing.T) {
	testCases := []struct {
		name    string
		order   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: ASC Order",
			order:   "DESC",
			wantErr: false,
		},
		{
			name:    "SUCCESS: DESC Order",
			order:   "ASC",
			wantErr: false,
		},
		{
			name:    "ERROR: JEDI Order",
			order:   "JEDI",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateSortOrder(tc.order)
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}

}
