package commonModel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMeta(t *testing.T) {

	testCases := []struct {
		Name               string
		Page               int64
		PerPage            int64
		TotalItems         int64
		ExpectedTotalPages int64
	}{
		{
			Name:               "SUCCESS: Correct total page for even number of items",
			Page:               1,
			PerPage:            10,
			TotalItems:         100,
			ExpectedTotalPages: 10,
		},
		{
			Name:               "SUCCESS: Correct total page for modulus not 0",
			Page:               1,
			PerPage:            10,
			TotalItems:         104,
			ExpectedTotalPages: 11,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			meta := NewMeta(testCase.Page, testCase.PerPage, testCase.TotalItems)
			assert.Equal(t, testCase.Page, meta.Page)
			assert.Equal(t, testCase.PerPage, meta.PerPage)
			assert.Equal(t, testCase.TotalItems, meta.TotalItems)
			if meta.TotalPages != testCase.ExpectedTotalPages {
				t.Errorf("Expected total pages to be %d but got %d", testCase.ExpectedTotalPages, meta.TotalPages)
			}
		})
	}
}
