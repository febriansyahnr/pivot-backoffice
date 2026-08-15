package account_model

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToResponse(t *testing.T) {
	accountUUID := uuid.New()
	merchantUUID := uuid.New()

	testCases := []struct {
		Name     string
		Input    *Account
		Expected *AccountResponse
	}{
		{
			Name: "it should return balance response",
			Input: &Account{
				UUID:        accountUUID,
				ReferenceID: merchantUUID,
				Name:        constant.TypeDisbursement,
				EODBalance:  10000.0,
				Currency:    constant.CurrencyIDR,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
				Type:        constant.TypeDisbursement,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &AccountResponse{
				UUID:        accountUUID,
				ReferenceID: merchantUUID,
				Name:        constant.TypeDisbursement,
				EODBalance:  10000.0,
				Currency:    constant.CurrencyIDR,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
				Type:        constant.TypeDisbursement,
				UserType:    constant.UserTypeMerchant,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Expected, tc.Input.ToResponse())
		})
	}
}

func TestToWalletResponse(t *testing.T) {
	accountUUID := uuid.New()
	merchantUUID := uuid.New()

	testCases := []struct {
		Name     string
		Input    *Account
		Expected *WalletAccountResponse
	}{
		{
			Name: "SUCCESS",
			Input: &Account{
				UUID:        accountUUID,
				ReferenceID: merchantUUID,
				Name:        constant.TypeWallet,
				EODBalance:  10000.0,
				Currency:    constant.CurrencyIDR,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeCustomer,
			},
			Expected: &WalletAccountResponse{
				UUID:        accountUUID,
				ReferenceID: merchantUUID,
				Name:        constant.TypeWallet,
				Currency:    constant.CurrencyIDR,
				CreatedAt:   util.TimeNow,
				UpdatedAt:   util.TimeNow,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			require.Equal(t, tc.Expected, tc.Input.ToWalletResponse())
		})
	}
}

func TestNewAccount(t *testing.T) {
	testcases := []struct {
		Name     string
		Input    *NewAccountRequest
		Expected *Account
		WantErr  bool
	}{
		{
			Name: "New merchant disbursement account",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceDisbursement,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeGeneralLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "New merchant disbursement account with type manual adjust",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypeManualAdjust,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeGeneralLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "New merchant disbursement account with type top up",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypeTopUp,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeGeneralLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "New SubMerchant disbursement account",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceDisbursement,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeSubMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "New SubMerchant payment account",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceDisbursement,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeSubMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Account Inquiry Usecase",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypeAccountInquiryFee,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeSubMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet Usecase",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeSubMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceWallet,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Customer Wallet Usecase ",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeCustomer,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.ReferenceWallet,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeCustomer,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Unknown Usecase ",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypeTopUp + constant.TypeAccountInquiryFee,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeCustomer,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.TypeDisbursement,
				EODBalance:  0,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeLedger,
				UserType:    constant.UserTypeCustomer,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Virtual Terminal",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypeVirtualTerminal,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.TypeVirtualTerminal,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeGeneralLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment Funded Payout",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.TypePaymentFundedPayout,
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: &Account{
				ReferenceID: uuid.Max,
				Name:        constant.TypePaymentFundedPayout,
				Currency:    constant.CurrencyIDR,
				Type:        constant.TypeGeneralLedger,
				UserType:    constant.UserTypeMerchant,
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Unknown UserType ",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     constant.ReferenceWallet,
				Currency:    constant.CurrencyIDR,
				UserType:    "Admin",
			},
			Expected: nil,
			WantErr:  true,
		},
		{
			Name: "ERROR: Invalid Usecase",
			Input: &NewAccountRequest{
				ReferenceID: uuid.Max,
				Usecase:     "INVALID_USECASE_THAT_RETURNS_EMPTY_STRING",
				Currency:    constant.CurrencyIDR,
				UserType:    constant.UserTypeMerchant,
			},
			Expected: nil,
			WantErr:  true,
		},
	}

	for _, test := range testcases {
		t.Run(test.Name, func(t *testing.T) {
			acc, err := NewAccount(test.Input)
			if test.WantErr {
				assert.NotNil(t, err)
				assert.Nil(t, acc)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, acc.ReferenceID, test.Expected.ReferenceID)
				assert.Equal(t, acc.Name, test.Expected.Name)
				assert.Equal(t, acc.Currency, test.Expected.Currency)
				assert.Equal(t, acc.Type, test.Expected.Type)
				assert.Equal(t, acc.UserType, test.Expected.UserType)

			}
		})
	}

}

func TestGetAccountName(t *testing.T) {
	testcases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Top Up Type Disbursement",
			Input:    constant.TypeTopUp,
			Expected: constant.TypeDisbursement,
		},
		{
			Name:     "Manual Adjust Type Disbursement",
			Input:    constant.TypeManualAdjust,
			Expected: constant.TypeDisbursement,
		},
		{
			Name:     "Payment",
			Input:    constant.TypePayment,
			Expected: constant.TypePayment,
		},
		{
			Name:     "Account Inquiry Fee",
			Input:    constant.TypeAccountInquiryFee,
			Expected: constant.TypeDisbursement,
		},
		{
			Name:     "Wallet",
			Input:    constant.TypeWallet,
			Expected: constant.TypeWallet,
		},
		{
			Name:     "Default",
			Input:    "a",
			Expected: "",
		},
	}

	for _, test := range testcases {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Expected, GetAccountName(test.Input))
		})
	}
}

func TestGetAccountNameByUsecase(t *testing.T) {
	testcases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Disbursement",
			Input:    constant.TypeDisbursement,
			Expected: constant.TypeDisbursement,
		},
		{
			Name:     "Payment",
			Input:    constant.TypePayment,
			Expected: constant.TypePayment,
		},
		{
			Name:     "Wallet",
			Input:    constant.TypeWallet,
			Expected: constant.TypeWallet,
		},
		{
			Name:     "Default",
			Input:    "a",
			Expected: constant.TypeDisbursement,
		},
	}

	for _, test := range testcases {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Expected, GetAccountNameByUsecase(test.Input))
		})
	}
}

func TestGetUserType(t *testing.T) {
	testcases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Merchant",
			Input:    constant.UserTypeMerchant,
			Expected: constant.UserTypeMerchant,
		},
		{
			Name:     "SubMerchant",
			Input:    constant.UserTypeSubMerchant,
			Expected: constant.UserTypeMerchant,
		},
		{
			Name:     "Customer",
			Input:    constant.UserTypeCustomer,
			Expected: constant.UserTypeCustomer,
		},
		{
			Name:     "Default",
			Input:    "a",
			Expected: "",
		},
	}

	for _, test := range testcases {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Expected, GetUserType(test.Input))
		})
	}
}

func TestGetLedgerType(t *testing.T) {
	testcases := []struct {
		Name     string
		Input    string
		Expected string
	}{
		{
			Name:     "Merchant",
			Input:    constant.UserTypeMerchant,
			Expected: constant.TypeGeneralLedger,
		},
		{
			Name:     "SubMerchant",
			Input:    constant.UserTypeSubMerchant,
			Expected: constant.TypeLedger,
		},
		{
			Name:     "Customer",
			Input:    constant.UserTypeCustomer,
			Expected: constant.TypeLedger,
		},
		{
			Name:     "Default",
			Input:    "a",
			Expected: "",
		},
	}

	for _, test := range testcases {
		t.Run(test.Name, func(t *testing.T) {
			require.Equal(t, test.Expected, GetLedgerType(test.Input))
		})
	}
}

func TestAccountRequiresPendingBalanceCalculation(t *testing.T) {
	tests := []struct {
		input      Account
		wantResult bool
	}{
		{
			input: Account{
				Name: constant.AccountNamePayment,
			},
			wantResult: true,
		},
		{
			input: Account{
				Name: constant.AccountNameWallet,
			},
			wantResult: true,
		},
		{
			input: Account{
				Name: constant.AccountNameVirtualTerminal,
			},
			wantResult: true,
		},
		{
			input: Account{
				Name: constant.AccountNameDisbursement,
			},
			wantResult: false,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.wantResult, tt.input.RequiresPendingBalanceCalculation())
	}
}

func TestAccountGetPendingTransactionCutoffOrBackdate(t *testing.T) {
	pendingTransactionCutoffAt := time.Now().UTC()

	tests := []struct {
		input      Account
		wantResult func(*testing.T, time.Time)
	}{
		{
			input: Account{},
			wantResult: func(t *testing.T, result time.Time) {
				assert.True(t, time.Now().UTC().AddDate(0, 0, -89).After(result))
			},
		},
		{
			input: Account{
				PendingTransactionCutoffAt: &pendingTransactionCutoffAt,
			},
			wantResult: func(t *testing.T, result time.Time) {
				assert.True(t, pendingTransactionCutoffAt.Equal(result))
			},
		},
	}

	for _, tt := range tests {
		result := tt.input.GetPendingTransactionCutoffOrBackdate()
		tt.wantResult(t, result)
	}
}
