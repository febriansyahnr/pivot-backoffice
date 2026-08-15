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

func TestCreatePayInTransactions(t *testing.T) {
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
			Name: "SUCCESS: Create Indirect Disbursement PayIn Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          merchantId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypeTopUp,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceDisbursement,
				TransactionTimestamp: time.Now(),
			},
			WantErr: false,
		},
		{
			Name: "SUCCESS: Create Disbursement Top Up to same recipient Requests",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceDisbursement,
				TransactionType:      constant.TypeTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          parentMerchantId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientAccountID:   parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypeTopUp,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceDisbursement,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment Request",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePayment,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          merchantAccountId,
				ParentID:             uuid.Max,
				ParentAccountID:      parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePayment,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment Request to main merchant",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePayment,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          parentMerchantId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientAccountID:   parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePayment,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet Top Up Request",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          merchantId,
				ParentID:             uuid.Max,
				ParentAccountID:      parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypeWalletTopUp,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Wallet Top Up Request to main merchant",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferenceWallet,
				TransactionType:      constant.TypeWalletTopUp,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          parentMerchantId,
				ParentID:             parentMerchantId,
				ParentAccountID:      parentMerchantAccountId,
				RecipientAccountID:   parentMerchantAccountId,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypeWalletTopUp,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferenceWallet,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment Request With Fee",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePayment,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          merchantAccountId,
				RecipientAccountID:   merchantAccountId,
				ParentID:             uuid.Max,
				ParentAccountID:      parentMerchantAccountId,
				MoneyFlowType:        constant.MoneyFlowIndirect,
				Fee: FeeRequest{
					Amount:             100,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePayment,
				TransactionTimestamp: time.Now(),
			},

			WantErr: false,
		},
		{
			Name: "SUCCESS: Payment Request Direct With Fee",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:          "123",
				Usecase:              constant.ReferencePayment,
				TransactionType:      constant.TypePayment,
				Channel:              constant.ChannelVirtualAccount,
				Remarks:              "test",
				TransactionTimestamp: time.Now(),
				Amount:               1000,
				Currency:             "IDR",
				TransferType:         constant.TransferTypePayIn,
				RecipientID:          merchantAccountId,
				RecipientAccountID:   merchantAccountId,
				ParentID:             uuid.Max,
				ParentAccountID:      parentMerchantAccountId,
				MoneyFlowType:        constant.MoneyFlowDirect,
				Fee: FeeRequest{
					Amount:             100,
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
				Channel:              constant.ChannelVirtualAccount,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Remarks:              "test",
				Reference:            constant.ReferencePayment,
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
				ParentID:             uuid.Max,
				ParentAccountID:      uuid.New(),
				RecipientAccountID:   uuid.New(),
				MoneyFlowType:        constant.MoneyFlowIndirect,
			},
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			trxList, err := CreatePayInTransactions(context.Background(), &tc.Request)
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
					if trx.Type != constant.TypeFee {
						assert.Equal(t, tc.ExpectedOutput.Channel, trx.Channel)
					}
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

				if tc.Request.MoneyFlowType == constant.MoneyFlowIndirect && tc.Request.RecipientID != tc.Request.ParentID {
					assert.Equal(t, tc.Request.Amount, trxList[0].Credit)
					assert.Equal(t, float64(0), trxList[0].Debit)
					assert.Equal(t, tc.Request.ParentAccountID, trxList[0].AccountID)

					assert.Equal(t, tc.Request.Amount, trxList[1].Debit)
					assert.Equal(t, float64(0), trxList[1].Credit)
					assert.Equal(t, tc.Request.ParentAccountID, trxList[1].AccountID)

					assert.Equal(t, tc.Request.Amount, trxList[2].Credit)
					assert.Equal(t, float64(0), trxList[2].Debit)
					assert.Equal(t, tc.Request.RecipientAccountID, trxList[2].AccountID)

					if tc.Request.Fee.Amount > 0 {
						assert.Equal(t, tc.Request.Fee.Amount, trxList[3].Debit)
						assert.Equal(t, float64(0), trxList[3].Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, trxList[3].AccountID)

						assert.Equal(t, tc.Request.Fee.Amount, trxList[4].Credit)
						assert.Equal(t, float64(0), trxList[4].Debit)
						assert.Equal(t, tc.Request.Fee.RecipientAccountID, trxList[4].AccountID)
					}

				} else {
					assert.Equal(t, tc.Request.Amount, trxList[0].Credit)
					assert.Equal(t, float64(0), trxList[0].Debit)
					assert.Equal(t, tc.Request.RecipientAccountID, trxList[0].AccountID)

					if tc.Request.Fee.Amount > 0 {
						assert.Equal(t, tc.Request.Fee.Amount, trxList[1].Debit)
						assert.Equal(t, float64(0), trxList[1].Credit)
						assert.Equal(t, tc.Request.RecipientAccountID, trxList[1].AccountID)

						assert.Equal(t, tc.Request.Fee.Amount, trxList[2].Credit)
						assert.Equal(t, float64(0), trxList[2].Debit)
						assert.Equal(t, tc.Request.Fee.RecipientAccountID, trxList[2].AccountID)
					}
				}
			}
		})
	}
}

func TestValidatePayInRequest(t *testing.T) {
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
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				MoneyFlowType:      constant.MoneyFlowIndirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "SUCCESS: Indirect money flow from parent",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.Max,
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.Max,
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
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
			ExpectedErr: nil,
		},
		{
			Name: "ERROR: Fee amount bigger than amount",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				ParentID:           uuid.New(),
				ParentAccountID:    uuid.New(),
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				Amount:             1000,
				Fee:                FeeRequest{Amount: 1001, RecipientAccountID: uuid.New()},
				MoneyFlowType:      constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrPayInFeeBiggerThanAmount,
		},
		{
			Name: "ERROR: Missing recipient ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:     "123",
				ParentID:        uuid.New(),
				ParentAccountID: uuid.New(),
				MoneyFlowType:   constant.MoneyFlowDirect,
			},
			ExpectedErr: constant.ErrMissingRecipientAccountID,
		},
		{
			Name: "ERROR: Missing Parent ID",
			Request: CreateNewLedgerEntryRequest{
				ReferenceID:        "123",
				RecipientID:        uuid.New(),
				RecipientAccountID: uuid.New(),
				MoneyFlowType:      constant.MoneyFlowIndirect,
			},
			ExpectedErr: constant.ErrMissingParentAccountID,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := tc.Request.ValidatePayInRequest()
			assert.Equal(t, tc.ExpectedErr, err)
		})
	}
}
