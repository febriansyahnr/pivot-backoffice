package user

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestValidateListUsersByMerchantIDRequest(t *testing.T) {

	testCases := []struct {
		Name     string
		Input    *ListUsersByMerchantIDRequest
		wantErr  bool
		Expected error
	}{
		{
			Name:     "SUCCESS",
			Input:    &ListUsersByMerchantIDRequest{},
			Expected: nil,
		},
		{
			Name: "SUCCESS: Sort by name",
			Input: &ListUsersByMerchantIDRequest{
				SortBy: "name",
			},
			Expected: nil,
		},
		{
			Name: "SUCCESS: Sort by name order",
			Input: &ListUsersByMerchantIDRequest{
				SortBy:    "name",
				SortOrder: "desc",
			},
			Expected: nil,
		},
		{
			Name: "ERROR: Sort by names",
			Input: &ListUsersByMerchantIDRequest{
				SortBy: "names",
			},
			wantErr:  true,
			Expected: constant.ErrInvalidUserListSortColumn,
		},
		{
			Name: "ERROR: Sort by name order",
			Input: &ListUsersByMerchantIDRequest{
				SortBy:    "name",
				SortOrder: "desca",
			},
			wantErr:  true,
			Expected: constant.ErrInvalidSortOrder,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Input.Validate()
			if tc.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.Expected.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
