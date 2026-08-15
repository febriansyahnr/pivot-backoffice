package snapCoreModel

import (
	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	"testing"
)

func TestVaTrxType(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want VirtualAccountTrxType
	}{
		{
			name: "Test Case 1",
			args: args{
				name: "1",
			},
			want: VirtualAccountTrxType{
				IsCloseAmount: false,
				IsSingleUsed:  false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := VaTrxType(tt.args.name); got != tt.want {
				t.Errorf("VaTrxType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindVaTrxTypeByCriteria(t *testing.T) {
	testCases := []struct {
		name          string
		isCloseAmount bool
		isSingleUsed  bool
		expected      string
	}{
		{
			name:          "Test Case 1: Open Static",
			isCloseAmount: false,
			isSingleUsed:  false,
			expected:      constant.VIRTUAL_ACCOUNT_TRX_TYPE_OPEN_STATIC,
		},
		{
			name:          "Test Case 2: Closed Static",
			isCloseAmount: true,
			isSingleUsed:  false,
			expected:      constant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_STATIC,
		},
		{
			name:          "Test Case 3: Closed Dynamic",
			isCloseAmount: true,
			isSingleUsed:  true,
			expected:      constant.VIRTUAL_ACCOUNT_TRX_TYPE_CLOSED_DYNAMIC,
		},
		{
			name:          "Test Case 4: No Match",
			isCloseAmount: false,
			isSingleUsed:  true,
			expected:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FindVaTrxTypeByCriteria(tc.isCloseAmount, tc.isSingleUsed)
			if result != tc.expected {
				t.Errorf("FindVaTrxTypeByCriteria() = %v, want %v", result, tc.expected)
			}
		})
	}
}
