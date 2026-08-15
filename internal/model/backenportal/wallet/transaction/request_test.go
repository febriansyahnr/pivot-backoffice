package walletTransactionModel_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/wallet/transaction"

	"github.com/stretchr/testify/assert"
)

func TestMerchantTransactionHistoryListReq(t *testing.T) {
	request := &MerchantTransactionHistoryListReq{
		MerchantId: "c7bef828-8269-490f-969d-3c5999218aeb",
		Type:       "MERCHANT_TOP_UP",
		Status:     "SUCCESS",
		Id:         "123456",
	}
	want := "b21764cfecd3ad641c9db0175fb64fdb8d28cbfadbc1fb719cfbaff123b5fb36"
	assert.Equal(t, want, request.HashFilter(constant.TimeLoc))

	request.EndDate = time.Now().UTC().Add(time.Hour)
	assert.NotEqual(t, want, request.HashFilter(constant.TimeLoc))
}
