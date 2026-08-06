package ledger_model

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	"github.com/stretchr/testify/assert"
)

func TestCreatePayOutTransactions(t *testing.T) {
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
			Name: "SUCCESS: Create Disbursement PayOut Requests from parent merchant",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				SenderID:             merchantId,
				SenderAccountID:      merchantAccountId,
				ParentID:             merchantId,
				ParentAccountID:      merchantAccountId,
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
				Channel:              constant.ChannelBankTransfer,
				Type:                 constant.TypeDisbursement,
				Status:               constant.StatusPending,
				Remarks:              "test",
				Reference:            constant.ReferenceDisbursement,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Disbursement PayOut Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				SenderID:             merchantId,
				SenderAccountID:      merchantAccountId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
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
				Channel:              constant.ChannelBankTransfer,
				Type:                 constant.TypeDisbursement,
				Status:               constant.StatusPending,
				Remarks:              "test",
				Reference:            constant.ReferenceDisbursement,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Wallet Withdrawal Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletWithdrawal,
				Channel:              constant.ChannelBankTransfer,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayOut,
				SenderID:             merchantId,
				SenderAccountID:      merchantAccountId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
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
				Channel:              constant.ChannelBankTransfer,
				Type:                 constant.TypeWalletWithdrawal,
				Status:               constant.StatusPending,
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
				TransferType:         constant.TransferTypePayOut,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientAccountID:   uuid.New(),
				RecipientID:          uuid.New(),
				MoneyFlowType:        constant.MoneyFlowIndirect,
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			trxList, err := CreatePayOutTransactions(context.Background(), &tc.Request)
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
					assert.Equal(t, tc.ExpectedOutput.Type, trx.Type)
					assert.NotNil(t, trx.TransactionTimestamp)
					assert.NotNil(t, trx.CreatedAt)
					assert.NotNil(t, trx.UpdatedAt)
				}

				assert.Equal(t, trxList[0].AccountID, tc.Request.SenderAccountID)
				assert.Equal(t, trxList[0].Debit, tc.Request.Amount)
				assert.Equal(t, trxList[0].Credit, float64(0))

				if tc.Request.Fee.Amount > 0 && tc.Request.Fee.RecipientAccountID != uuid.Nil {
					assert.Equal(t, trxList[1].AccountID, tc.Request.SenderAccountID)
					assert.Equal(t, trxList[1].Debit, tc.Request.Fee.Amount)
					assert.Equal(t, trxList[1].Credit, float64(0))

					assert.Equal(t, trxList[2].AccountID, tc.Request.Fee.RecipientAccountID)
					assert.Equal(t, trxList[2].Credit, tc.Request.Fee.Amount)
					assert.Equal(t, trxList[2].Debit, float64(0))

					if tc.Request.MoneyFlowType == constant.MoneyFlowIndirect &&
						tc.Request.SenderID != tc.Request.ParentID {
						assert.Equal(t, trxList[3].AccountID, tc.Request.ParentAccountID)
						assert.Equal(t, trxList[3].Credit, tc.Request.Amount)
						assert.Equal(t, trxList[3].Debit, float64(0))

						assert.Equal(t, trxList[4].AccountID, tc.Request.ParentAccountID)
						assert.Equal(t, trxList[4].Debit, tc.Request.Amount)
						assert.Equal(t, trxList[4].Credit, float64(0))
					}
				} else {
					if tc.Request.MoneyFlowType == constant.MoneyFlowIndirect &&
						tc.Request.SenderID != tc.Request.ParentID {
						assert.Equal(t, trxList[1].AccountID, tc.Request.ParentAccountID)
						assert.Equal(t, trxList[1].Credit, tc.Request.Amount)
						assert.Equal(t, trxList[1].Debit, float64(0))

						assert.Equal(t, trxList[2].AccountID, tc.Request.ParentAccountID)
						assert.Equal(t, trxList[2].Debit, tc.Request.Amount)
						assert.Equal(t, trxList[2].Credit, float64(0))
					}
				}

			}
		})
	}
}

func TestValidatePayOutRequest(t *testing.T) {
	testCases := []struct {
		Name        string
		Request     CreateNewLedgerEntryRequest
		ExpectedErr error
	}{
		{
			Name: "SUCCESS: Indirect money flow",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				SenderID:        uuid.New(),
				SenderAccountID: uuid.New(),
				ParentID:        uuid.New(),
				ParentAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowIndirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "SUCCESS: Indirect money flow from parent",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				SenderID:        uuid.New(),
				SenderAccountID: uuid.New(),
				ParentID:        uuid.New(),
				ParentAccountID: uuid.Max,
				MoneyFlowType:   constant.MoneyFlowIndirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "SUCCESS: Direct money flow",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				SenderID:        uuid.New(),
				SenderAccountID: uuid.New(),
				ParentID:        uuid.New(),
				ParentAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowDirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "ERROR: Missing sender ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				ParentID:        uuid.New(),
				ParentAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrMissingSenderAccountID,
		},
		{
			Name: "ERROR: Missing Parent ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				SenderID:        uuid.New(),
				SenderAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowIndirect,
			},
			ExpectedErr: constant.ErrMissingParentAccountID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.ValidatePayOutRequest()
			assert.Equal(t, tc.ExpectedErr, err)
		})
	}
}
