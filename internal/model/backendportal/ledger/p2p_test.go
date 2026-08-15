package ledger_model

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestCreateP2PTransactions(t *testing.T) {

	senderID := uuid.New()
	SenderAccountID := uuid.New()
	merchantId := uuid.New()
	merchantAccountId := uuid.New()
	parentMerchantId := uuid.New()
	parentMerchantAccountId := uuid.New()

	testCases := []struct {
		Name           string
		Request        CreateNewLedgerEntryRequest
		ExpectedOutput orchestrator_model.AccountTransaction
		WantErr        bool
	}{
		{
			Name: "SUCCESS: Create P2P Requests With Type Manual Topup",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.WalletTrxTopUpType,
				Channel:              constant.ChannelManualTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				RecipientID:          merchantId,
				RecipientAccountID:   merchantAccountId,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: FeeRequest{
					Amount:             1000,
					Channel:            constant.ChannelManualTransfer,
					RecipientID:        parentMerchantId,
					RecipientAccountID: parentMerchantAccountId,
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
				Channel:              constant.ChannelManualTransfer,
				Type:                 constant.WalletTrxTopUpType,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create P2P Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				SenderID:             parentMerchantId,
				SenderAccountID:      parentMerchantAccountId,
				RecipientID:          merchantId,
				RecipientAccountID:   merchantAccountId,
				MoneyFlowType:        constant.MoneyFlowIndirect,
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
				Type:                 constant.TypeWalletTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create P2P Requests via Parent Merchant",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				SenderID:             senderID,
				SenderAccountID:      SenderAccountID,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientID:          merchantId,
				RecipientAccountID:   merchantAccountId,
				MoneyFlowType:        constant.MoneyFlowIndirect,
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
				Type:                 constant.TypeWalletTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create P2P Requests via Parent Merchant with Fee",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTransfer,
				Channel:              "",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypeP2P,
				SenderID:             senderID,
				SenderAccountID:      SenderAccountID,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientID:          merchantId,
				RecipientAccountID:   merchantAccountId,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: FeeRequest{
					Amount:             100,
					RecipientID:        parentMerchantId,
					RecipientAccountID: parentMerchantAccountId,
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
				Type:                 constant.TypeWalletTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "ERROR: Invalid Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement + "invalid",
				TransactionType:      constant.TypeTopUp,
				Channel:              "CHANNEL",
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          uuid.New(),
				MoneyFlowType:        constant.MoneyFlowIndirect,
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			trxList, err := CreateP2PTransactions(context.Background(), &tc.Request)
			if tc.WantErr {
				assert.Nil(t, trxList)
				assert.NotNil(t, err)

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
				assert.Equal(t, tc.Request.SenderAccountID, debitTrx.AccountID)

				if tc.Request.Fee.Amount > 0 {
					feeDebitTrx := trxList[1]
					assert.Equal(t, float64(0), feeDebitTrx.Credit)
					assert.Equal(t, tc.Request.Fee.Amount, feeDebitTrx.Debit)
					if tc.Request.TransactionType == constant.WalletTrxTopUpType {
						assert.Equal(t, tc.Request.RecipientAccountID, feeDebitTrx.AccountID)
						if tc.Request.Fee.RecipientAccountID != uuid.Nil {
							assert.Equal(t, tc.Request.SenderAccountID, trxList[2].AccountID)
						}
					} else {
						assert.Equal(t, tc.Request.SenderAccountID, feeDebitTrx.AccountID)
					}

					recipientFeeCreditTrx := trxList[2]
					assert.Equal(t, float64(0), recipientFeeCreditTrx.Debit)
					assert.Equal(t, tc.Request.Fee.Amount, recipientFeeCreditTrx.Credit)
					assert.Equal(t, tc.Request.Fee.RecipientAccountID, recipientFeeCreditTrx.AccountID)

					if tc.Request.ParentAccountID != uuid.Nil &&
						tc.Request.ParentAccountID != tc.Request.SenderAccountID &&
						tc.Request.ParentAccountID != tc.Request.RecipientAccountID {
						glCreditTrx := trxList[3]
						assert.Equal(t, float64(0), glCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, glCreditTrx.Credit)
						assert.Equal(t, tc.Request.ParentAccountID, glCreditTrx.AccountID)

						glDebitTrx := trxList[4]
						assert.Equal(t, float64(0), glDebitTrx.Credit)
						assert.Equal(t, tc.Request.Amount, glDebitTrx.Debit)
						assert.Equal(t, tc.Request.ParentAccountID, glDebitTrx.AccountID)

						recipientCreditTrx := trxList[5]
						assert.Equal(t, float64(0), recipientCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, recipientCreditTrx.Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, recipientCreditTrx.AccountID)
					} else {
						recipientCreditTrx := trxList[3]
						assert.Equal(t, float64(0), recipientCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, recipientCreditTrx.Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, recipientCreditTrx.AccountID)

					}
				} else {

					if tc.Request.ParentAccountID != uuid.Nil &&
						tc.Request.ParentAccountID != tc.Request.SenderAccountID &&
						tc.Request.ParentAccountID != tc.Request.RecipientAccountID {
						glCreditTrx := trxList[1]
						assert.Equal(t, float64(0), glCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, glCreditTrx.Credit)
						assert.Equal(t, tc.Request.ParentAccountID, glCreditTrx.AccountID)

						glDebitTrx := trxList[2]
						assert.Equal(t, float64(0), glDebitTrx.Credit)
						assert.Equal(t, tc.Request.Amount, glDebitTrx.Debit)
						assert.Equal(t, tc.Request.ParentAccountID, glDebitTrx.AccountID)

						recipientCreditTrx := trxList[3]
						assert.Equal(t, float64(0), recipientCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, recipientCreditTrx.Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, recipientCreditTrx.AccountID)

					} else {
						recipientCreditTrx := trxList[1]
						assert.Equal(t, float64(0), recipientCreditTrx.Debit)
						assert.Equal(t, tc.Request.Amount, recipientCreditTrx.Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, recipientCreditTrx.AccountID)
					}
				}
			}
		})
	}
}

func TestValidateP2PRequest(t *testing.T) {
	testCases := []struct {
		Name        string
		Request     CreateNewLedgerEntryRequest
		ExpectedErr error
	}{
		{
			Name: "SUCCESS: Indirect money flow",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				SenderID:           uuid.New(),
				SenderAccountID:    uuid.New(),
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				MoneyFlowType:      constant.MoneyFlowIndirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "SUCCESS: Direct money flow",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				SenderID:           uuid.New(),
				SenderAccountID:    uuid.New(),
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "ERROR: Missing recipient ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				SenderID:        uuid.New(),
				SenderAccountID: uuid.New(),
				ParentID:        uuid.New(),
				ParentAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrMissingRecipientAccountID,
		},
		{
			Name: "ERROR: Same recipient & sender ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				SenderID:           uuid.Max,
				SenderAccountID:    uuid.Max,
				RecipientID:        uuid.Max,
				RecipientAccountID: uuid.Max,
				ParentID:           uuid.Max,
				ParentAccountID:    uuid.Max,
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrSenderSameWithRecipient,
		},
		{
			Name: "ERROR: Missing sender ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrMissingSenderAccountID,
		},
		{
			Name: "ERROR: Missing Parent ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				SenderID:           uuid.New(),
				SenderAccountID:    uuid.New(),
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				MoneyFlowType:      constant.MoneyFlowIndirect,
			},
			ExpectedErr: constant.ErrParentAccountNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.ValidateP2PRequest()
			assert.Equal(t, tc.ExpectedErr, err)
		})
	}
}
