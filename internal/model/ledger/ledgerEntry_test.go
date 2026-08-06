package ledger_model

import (
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/require"
)

func TestValidateLedgerEntryCreation(t *testing.T) {

	testCases := []struct {
		Name    string
		Request CreateNewLedgerEntryRequest
		WantErr bool
	}{
		{
			Name: "SUCCESS: Valid P2P Transfer Type",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypeP2P,
				Usecase:      constant.ReferenceDisbursement,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Valid PayIn Transfer Type",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayIn,
				Usecase:      constant.ReferenceDisbursement,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Valid PayOut Transfer Type",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceDisbursement,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Valid Disbursement Usecase",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceDisbursement,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Valid Payment Usecase",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferencePayment,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Valid Wallet Usecase",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceWallet,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Fee amount and recipient not empty",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceDisbursement,
				Amount:       10,
				Fee: FeeRequest{
					Amount:             1,
					RecipientAccountID: uuid.New(),
				},
			},
		},
		{
			Name: "SUCCESS: Fee recipient is not set",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceDisbursement,
				Amount:       10,
				Fee: FeeRequest{
					Amount:             2,
					RecipientAccountID: uuid.Nil,
				},
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Invalid Transfer Type",
			Request: CreateNewLedgerEntryRequest{
				TransferType: "",
				Usecase:      constant.ReferenceDisbursement,
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid use case",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypeP2P,
				Usecase:      constant.ReferenceAccountInquiry,
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Fee amount is less than 0",
			Request: CreateNewLedgerEntryRequest{
				TransferType: constant.TransferTypePayOut,
				Usecase:      constant.ReferenceDisbursement,
				Amount:       10,
				Fee: FeeRequest{
					Amount:             -1,
					RecipientAccountID: uuid.New(),
				},
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.Validate()
			if tc.WantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateTransferType(t *testing.T) {
	testCases := []struct {
		name         string
		transferType string
		wantErr      bool
	}{
		{
			name:         "SUCCESS: Valid Transfer Type P2P",
			transferType: constant.TransferTypeP2P,
			wantErr:      false,
		},
		{
			name:         "SUCCESS: Valid Transfer Type PayIn",
			transferType: constant.TransferTypePayIn,
			wantErr:      false,
		},
		{
			name:         "SUCCESS: Valid Transfer Type PayOut",
			transferType: constant.TransferTypePayOut,
			wantErr:      false,
		},
		{
			name:         "ERROR: Invalid Transfer Type",
			transferType: "invalid",
			wantErr:      true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTransferType(tc.transferType)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateUseCase(t *testing.T) {
	testCases := []struct {
		name    string
		usecase string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Valid Use Case Disbursement",
			usecase: constant.ReferenceDisbursement,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Valid Use Case Payment",
			usecase: constant.ReferencePayment,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Valid Use Case Wallet",
			usecase: constant.ReferenceWallet,
			wantErr: false,
		},
		{
			name:    "ERROR: Invalid Use Case",
			usecase: "invalid",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateUseCase(tc.usecase)
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
