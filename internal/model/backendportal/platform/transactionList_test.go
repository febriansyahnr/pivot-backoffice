package platform

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	testCases := []struct {
		name    string
		request TransactionRequest
		wantErr bool
	}{
		{
			name: "SUCCESS: DISBURSEMENT",
			request: TransactionRequest{
				Reference: constant.TypeDisbursement,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: PAYMENT",
			request: TransactionRequest{
				Reference: constant.TypePayment,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: PLATFORM TRANSFER",
			request: TransactionRequest{
				Reference: constant.ReferencePlatformTransfer,
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Valid Sort Order",
			request: TransactionRequest{
				Reference: constant.ReferencePlatformTransfer,
				SortOrder: "ASC",
			},
			wantErr: false,
		},

		{
			name: "ERROR: Unknwon Reference",
			request: TransactionRequest{
				Reference: constant.TypeXB,
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid Sort Order",
			request: TransactionRequest{
				Reference: constant.ReferencePlatformTransfer,
				SortOrder: "ASCs",
			},
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.request.Validate()
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
