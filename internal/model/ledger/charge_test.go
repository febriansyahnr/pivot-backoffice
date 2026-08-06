package ledger_model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestCreateChargeTransactions(t *testing.T) {
	merchantId := uuid.New()
	merchantAccountId := uuid.New()
	parentMerchantId := uuid.New()
	parentMerchantAccountId := uuid.New()
	feeRecipientId := uuid.New()
	feeRecipientAccountId := uuid.New()

	testCases := []struct {
		Name           string
		Request        CreateNewLedgerEntryRequest
		ExpectedOutput orchestrator_model.AccountTransaction
		WantErr        bool
	}{
		{
			Name: "SUCCESS: Create Charge Requests with Recipient",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				RecipientID:          merchantId,
				RecipientAccountID:   merchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			ExpectedOutput: orchestrator_model.AccountTransaction{
				UUID:                 uuid.New(),
				ReferenceID:          "123",
				MerchantID:           merchantId,
				AccountID:            merchantAccountId,
				Currency:             "IDR",
				Credit:               1000,
				Debit:                0,
				Channel:              "",
				Type:                 constant.TypeFee,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePlatform,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Charge Request",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			ExpectedOutput: orchestrator_model.AccountTransaction{
				UUID:                 uuid.New(),
				ReferenceID:          "123",
				MerchantID:           merchantId,
				AccountID:            merchantAccountId,
				Currency:             "IDR",
				Credit:               1000,
				Debit:                0,
				Channel:              "",
				Type:                 constant.TypeFee,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePlatform,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},

		{
			Name: "SUCCESS: Create Charge Request Indirectly Deducted",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: false,
				},
			},
			ExpectedOutput: orchestrator_model.AccountTransaction{
				UUID:                 uuid.New(),
				ReferenceID:          "123",
				MerchantID:           merchantId,
				AccountID:            merchantAccountId,
				Currency:             "IDR",
				Credit:               1000,
				Debit:                0,
				Channel:              "",
				Type:                 constant.TypeFee,
				Status:               constant.StatusPending,
				Remarks:              "test",
				Reference:            constant.ReferencePlatform,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Charge Request with Fee",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				Fee: FeeRequest{
					Amount:             100,
					RecipientID:        feeRecipientId,
					RecipientAccountID: feeRecipientAccountId,
					AdditionalInfo:     map[string]interface{}{"fee_type": "service"},
				},
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			ExpectedOutput: orchestrator_model.AccountTransaction{
				UUID:                 uuid.New(),
				ReferenceID:          "123",
				MerchantID:           merchantId,
				AccountID:            merchantAccountId,
				Currency:             "IDR",
				Credit:               1000,
				Debit:                0,
				Channel:              "",
				Type:                 constant.TypeFee,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePlatform,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Charge Request with Fee but no Recipient",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				Fee: FeeRequest{
					Amount:         100,
					AdditionalInfo: map[string]interface{}{"fee_type": "service"},
					// No RecipientID or RecipientAccountID
				},
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			ExpectedOutput: orchestrator_model.AccountTransaction{
				UUID:                 uuid.New(),
				ReferenceID:          "123",
				MerchantID:           merchantId,
				AccountID:            merchantAccountId,
				Currency:             "IDR",
				Credit:               1000,
				Debit:                0,
				Channel:              "",
				Type:                 constant.TypeFee,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePlatform,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Invalid Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement + "s",
				TransactionType:      constant.TypeTopUp,
				Channel:              "CHANNEL",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          uuid.New(),
				MoneyFlowType:        constant.MoneyFlowIndirect,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Missing SenderID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             uuid.Nil,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Missing SenderAccountID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      uuid.Nil,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Empty Currency",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Amount",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               0,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Zero Transaction Timestamp",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Time{},
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Fee Transaction",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				Fee: FeeRequest{
					Amount:             100,
					RecipientID:        feeRecipientId,
					RecipientAccountID: feeRecipientAccountId,
					AdditionalInfo:     "invalid", // Invalid AdditionalInfo type
				},
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Status in Main Transaction",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              "INVALID_USECASE", // This will cause NewAccountTransaction to fail
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Fee Transaction with Invalid AdditionalInfo",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePlatform,
				TransactionType:      constant.TypeFee,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeCharge,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				Fee: FeeRequest{
					Amount:             100,
					RecipientID:        feeRecipientId,
					RecipientAccountID: feeRecipientAccountId,
					AdditionalInfo:     func() {}, // Invalid AdditionalInfo type that will fail JSON marshaling
				},
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Invalid Recipient Transaction",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:             "123",
				Usecase:                 constant.ReferencePlatform,
				TransactionType:         constant.TypeFee,
				Channel:                 "",
				Remarks:                 "test",
				TransactionTimestamp:    time.Now(),
				Amount:                  1000,
				Currency:                "IDR",
				TransferType:            constant.TransferTypeCharge,
				SenderID:                parentMerchantId,
				SenderAccountID:         parentMerchantAccountId,
				RecipientID:             merchantId,
				RecipientAccountID:      merchantAccountId,
				SenderAdditionalInfo:    func() {}, // Invalid AdditionalInfo type that will fail JSON marshaling
				RecipientAdditionalInfo: func() {}, // Invalid AdditionalInfo type that will fail JSON marshaling
				ChargeConfig: ChargeConfig{
					BypassBalanceCheck: true,
					IsDirectlyDeducted: true,
				},
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			trxList, err := CreateChargeTransactions(context.Background(), &tc.Request)
			if tc.WantErr {
				assert.Nil(t, trxList)
				assert.NotNil(t, err)

				// Check for specific error when SenderID or SenderAccountID is nil
				if tc.Name == "ERROR: Missing SenderID" || tc.Name == "ERROR: Missing SenderAccountID" {
					assert.Equal(t, constant.ErrMissingSenderAccountID, err, "Expected ErrMissingSenderAccountID error")
				}

				// Check for specific error when currency is empty
				if tc.Name == "ERROR: Empty Currency" {
					assert.Equal(t, "invalid currency", err.Error(), "Expected 'invalid currency' error")
				}

				// Check for specific error when amount is invalid
				if tc.Name == "ERROR: Invalid Amount" {
					assert.Equal(t, constant.ErrInvalidAmount, err, "Expected ErrInvalidAmount error")
				}

				// Check for specific error when transaction timestamp is zero
				if tc.Name == "ERROR: Zero Transaction Timestamp" {
					assert.Equal(t, constant.ErrInvalidTransactionTimestamp, err, "Expected ErrInvalidTransactionTimestamp error")
				}

				// Check for specific error when fee AdditionalInfo is invalid
				if tc.Name == "ERROR: Invalid Fee Transaction" {
					assert.Equal(t, "invalid fee additional info type", err.Error(), "Expected 'invalid fee additional info type' error")
				}

				// Check for errors from NewAccountTransaction
				if tc.Name == "ERROR: Invalid Status in Main Transaction" ||
					tc.Name == "ERROR: Invalid Channel in Main Transaction" ||
					tc.Name == "ERROR: Invalid Fee Transaction with Invalid AdditionalInfo" ||
					tc.Name == "ERROR: Invalid Recipient Transaction" {
					assert.NotNil(t, err, "Expected error from NewAccountTransaction")
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, trxList)

				for _, trx := range trxList {
					assert.NotNil(t, trx.UUID)
					assert.NotNil(t, trx.AccountID)
					assert.NotNil(t, trx.MerchantID)
					assert.Equal(t, tc.ExpectedOutput.Channel, trx.Channel)
					assert.Equal(t, tc.ExpectedOutput.Currency, trx.Currency)
					assert.Equal(t, tc.ExpectedOutput.ReasonDescription, trx.ReasonDescription)
					assert.Equal(t, tc.ExpectedOutput.ReasonType, trx.ReasonType)
					assert.Equal(t, tc.ExpectedOutput.Reference, trx.Reference)
					assert.Equal(t, tc.ExpectedOutput.ReferenceID, trx.ReferenceID)
					assert.Equal(t, tc.ExpectedOutput.Remarks, trx.Remarks)
					assert.Equal(t, tc.ExpectedOutput.Status, trx.Status)
					assert.NotNil(t, tc.ExpectedOutput.Type)
					assert.NotNil(t, trx.TransactionTimestamp)
					assert.NotNil(t, trx.CreatedAt)
					assert.NotNil(t, trx.UpdatedAt)
				}

				debitTrx := trxList[0]
				assert.Equal(t, float64(0), debitTrx.Credit)
				assert.NotEqual(t, float64(0), debitTrx.Debit)
				assert.Equal(t, tc.Request.Amount, debitTrx.Debit)
				assert.Equal(t, tc.Request.SenderAccountID, debitTrx.AccountID)

				// Check fee transactions if they exist
				if tc.Request.Fee.Amount > 0 {
					// Check if we have at least 2 transactions (main debit + fee debit)
					assert.GreaterOrEqual(t, len(trxList), 2, "Should have at least 2 transactions when fee is present")

					// Check fee debit transaction
					feeDebitTrx := trxList[1]
					assert.Equal(t, float64(0), feeDebitTrx.Credit)
					assert.Equal(t, tc.Request.Fee.Amount, feeDebitTrx.Debit)
					assert.Equal(t, tc.Request.SenderAccountID, feeDebitTrx.AccountID)
					assert.Equal(t, constant.TypeFee, feeDebitTrx.Type)

					// Check fee credit transaction if recipient is specified
					if tc.Request.Fee.RecipientAccountID != uuid.Nil {
						assert.Equal(t, 3, len(trxList), "Should have 3 transactions when fee recipient is present")
						feeCreditTrx := trxList[2]
						assert.Equal(t, float64(0), feeCreditTrx.Debit)
						assert.Equal(t, tc.Request.Fee.Amount, feeCreditTrx.Credit)
						assert.Equal(t, tc.Request.Fee.RecipientAccountID, feeCreditTrx.AccountID)
						assert.Equal(t, constant.TypeFee, feeCreditTrx.Type)
					} else {
						// If no recipient is specified, we should only have 2 transactions
						assert.Equal(t, 2, len(trxList), "Should have 2 transactions when fee recipient is not present")
					}
				}

				// Check recipient credit transaction if it exists
				if tc.Request.RecipientAccountID != uuid.Nil {
					// Find the credit transaction for the recipient
					var recipientCreditTrx *orchestrator_model.AccountTransaction
					for _, trx := range trxList {
						if trx.AccountID == tc.Request.RecipientAccountID && trx.Credit > 0 {
							recipientCreditTrx = trx
							break
						}
					}

					assert.NotNil(t, recipientCreditTrx, "Should have a credit transaction for recipient")
					assert.Equal(t, float64(0), recipientCreditTrx.Debit)
					assert.Equal(t, tc.Request.Amount, recipientCreditTrx.Credit)
					assert.Equal(t, tc.Request.RecipientAccountID, recipientCreditTrx.AccountID)
				}
			}
		})
	}
}

func TestCreateChargeTransactionsUUIDError(t *testing.T) {
	// Save the original generator from orchestrator model and restore it after the test
	originalGenerator := orchestrator_model.GetDefaultUUIDGenerator()
	defer func() {
		orchestrator_model.SetDefaultUUIDGenerator(originalGenerator)
	}()

	// Replace the UUID generator in the orchestrator model with one that returns an error
	orchestrator_model.SetDefaultUUIDGenerator(func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("simulated UUID generation error")
	})

	// Create a basic valid request
	request := &CreateNewLedgerEntryRequest{
		ReferenceID:          "123",
		Usecase:              constant.ReferencePlatform,
		TransactionType:      constant.TypeFee,
		Channel:              "",
		Remarks:              "test",
		TransactionTimestamp: time.Now(),
		Amount:               1000,
		Currency:             "IDR",
		TransferType:         constant.TransferTypeCharge,
		SenderID:             uuid.New(),
		SenderAccountID:      uuid.New(),
		ChargeConfig: ChargeConfig{
			BypassBalanceCheck: true,
			IsDirectlyDeducted: true,
		},
	}

	// Call the function with the mock generator
	result, err := CreateChargeTransactions(context.Background(), request)

	// Verify that the error is propagated and the result is nil
	assert.Error(t, err, "Function should return an error when UUID generation fails")
	assert.Nil(t, result, "Result should be nil when an error occurs")
	assert.Contains(t, err.Error(), "simulated UUID generation error", "Error message should be propagated")
}
