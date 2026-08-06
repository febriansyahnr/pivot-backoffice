package ledger_model

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/stretchr/testify/assert"
)

func TestNewRefundTransactions(t *testing.T) {
	mockTime := time.Now()

	tests := []struct {
		name          string
		request       *CreateNewLedgerEntryRequest
		expectedCount int
		expectedError bool
	}{
		{
			name: "success with refund to sender first and fee",
			request: &CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-1",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				MerchantReferenceID:  &[]string{"merchant-ref-1"}[0],
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				Remarks:              "Test refund",
				TransactionTimestamp: mockTime,
				Fee: FeeRequest{
					Amount:             50.0,
					TransactionType:    constant.TypeFeeReversal,
					Channel:            "FEE",
					RecipientID:        uuid.New(),
					RecipientAccountID: uuid.New(),
				},
				RefundConfig: RefundConfig{
					RefundToSenderFirst: true,
				},
			},
			expectedCount: 4,
			expectedError: false,
		},
		{
			name: "success without refund to sender first but with fee",
			request: &CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-2",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				Remarks:              "Test refund",
				TransactionTimestamp: mockTime,
				Fee: FeeRequest{
					Amount:             50.0,
					TransactionType:    constant.TypeFeeReversal,
					Channel:            "FEE",
					RecipientID:        uuid.New(),
					RecipientAccountID: uuid.New(),
				},
				RefundConfig: RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			expectedCount: 3,
			expectedError: false,
		},
		{
			name: "success without fee",
			request: &CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-3",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               1000.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				Remarks:              "Test refund",
				TransactionTimestamp: mockTime,
				Fee: FeeRequest{
					Amount: 0.0,
				},
				RefundConfig: RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			expectedCount: 2,
			expectedError: false,
		},
		{
			name: "minimal configuration",
			request: &CreateNewLedgerEntryRequest{
				ReferenceID:          "test-ref-4",
				SenderID:             uuid.New(),
				SenderAccountID:      uuid.New(),
				RecipientID:          uuid.New(),
				RecipientAccountID:   uuid.New(),
				Amount:               500.0,
				TransactionType:      constant.TypeRefund,
				Channel:              "TEST",
				Usecase:              constant.ReferenceWallet,
				TransactionTimestamp: mockTime,
				Fee: FeeRequest{
					Amount: 0.0,
				},
				RefundConfig: RefundConfig{
					RefundToSenderFirst: false,
				},
			},
			expectedCount: 2,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := NewRefundTransactions(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, tt.expectedCount)

			// Verify each transaction has required fields
			for _, trx := range result {
				assert.NotEmpty(t, trx.ReferenceID)
				assert.NotEmpty(t, trx.Currency)
				assert.Equal(t, constant.StatusSuccess, trx.Status)
			}
		})
	}
}
