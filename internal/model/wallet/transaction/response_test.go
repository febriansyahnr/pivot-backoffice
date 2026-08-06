package walletTransactionModel_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"

	"github.com/stretchr/testify/assert"
)

func TestTransactionTypeToTitle(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{
			input: "FEE",
			want:  "Fee",
		},
		{
			input: "FEE_WALLET_TRANSACTION",
			want:  "Wallet Transaction Fee",
		},
		{
			input: "MERCHANT_PAYMENT",
			want:  "Merchant Payment",
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.want, TransactionTypeToTitle(test.input))
	}
}

func TestMerchantTransactionHistoryListResp(t *testing.T) {
	response := MerchantTransactionHistoryListResp{
		Type:    constant.TypeWithdrawal,
		Channel: constant.ChannelBankTransfer,
	}
	assert.Equal(t, "Withdrawal", response.TransactionTypeToTitle())
	assert.Equal(t, "-", response.GetReferenceID())

	response.Type = constant.WalletTrxTopUpType
	response.Channel = constant.ChannelManualTransfer
	response.ReferenceId = "REF/****" // NOSONAR
	assert.Equal(t, "Top Up", response.TransactionTypeToTitle())
	assert.Equal(t, "REF/****", response.GetReferenceID()) // NOSONAR
}
