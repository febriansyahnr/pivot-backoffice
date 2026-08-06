package paymentModel_test

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestCalculateInvestigationMonthlyReconciliation(t *testing.T) {

	id := util.GenerateUUID()
	date := time.Now().UTC()

	transaction := &CalculateInvestigationMonthlyReconciliation{
		MerchantID:             "0fbb7e68-e08a-4b51-977d-336762c955b3",
		RawPaymentIDs:          []byte(`["08424aed-8cb6-465b-9c62-c8422b0778f7"]`),
		PaymentIDs:             []string{"08424aed-8cb6-465b-9c62-c8422b0778f7"},
		PaymentCount:           1,
		GrossAmount:            50_000,
		FeeAmount:              2_500,
		NetAmount:              47_500,
		PlatformLossPercentage: 10,
		PlatformMaxLoss:        100_000,
		PlatformLossAmount:     4_750,
		MerchantLossAmount:     42_750,
	}

	t.Run("To create account transaction request", func(t *testing.T) {
		result := transaction.ToCreateAccountTransactionRequest(id.String(), date)

		wantResult := &orchestrator_model.CreateAccountTransactionRequest{
			UUID:                 result.UUID,
			ReferenceID:          id.String(),
			MerchantID:           util.ParseUUID(transaction.MerchantID),
			Currency:             constant.CurrencyIDR,
			Debit:                transaction.MerchantLossAmount,
			Type:                 constant.TypePayment,
			Channel:              constant.ChannelInvestigation,
			Status:               constant.StatusSuccess,
			TransactionTimestamp: date,
			ReasonType:           util.ValueToPtr(constant.TypeFinalFailedDeduction),
			SettlementModel:      util.ValueToPtr(constant.SettlementModelAggregator),
			Usecase:              constant.TypePayment,
		}
		assert.Equal(t, wantResult, result)
	})

	t.Run("To payment investigation monthly reconciliation", func(t *testing.T) {
		result := transaction.ToPaymentInvestigationMonthlyReconciliation(id.String(), date)

		wantResult := PaymentInvestigationMonthlyReconciliation{
			UUID:                   id.String(),
			Date:                   date,
			MerchantID:             transaction.MerchantID,
			RawPaymentIDs:          transaction.RawPaymentIDs,
			PaymentIDs:             transaction.PaymentIDs,
			PaymentCount:           transaction.PaymentCount,
			GrossAmount:            transaction.GrossAmount,
			FeeAmount:              transaction.FeeAmount,
			NetAmount:              transaction.NetAmount,
			PlatformLossPercentage: transaction.PlatformLossPercentage,
			PlatformMaxLoss:        transaction.PlatformMaxLoss,
			PlatformLossAmount:     transaction.PlatformLossAmount,
			MerchantLossAmount:     transaction.MerchantLossAmount,
			CreatedAt:              result.CreatedAt,
		}
		assert.Equal(t, wantResult, result)
	})
}
