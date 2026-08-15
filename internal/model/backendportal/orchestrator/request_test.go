package orchestrator_model_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	account_model "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/account"
	. "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/orchestrator"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestModelCreateAccountTransactionRequest(pt *testing.T) {
	id, _ := uuid.Parse("e39d1d21-0a7b-49e5-8cd9-404ac75d54be")

	accountTransaction := &CreateAccountTransactionRequest{
		UUID:        id,
		MerchantID:  id,
		ReferenceID: "REF-0001",
		Currency:    "IDR",
	}

	routing := map[string]string{
		"":                             "",
		constant.ChannelBalance:        constant.TypeDisbursement,
		constant.ChannelVirtualAccount: constant.TypePayment,
		constant.ChannelCreditCard:     constant.TypePayment,
		constant.ChannelBankTransfer:   constant.TypePayment,
	}
	pt.Run("RUN:Routing channel", func(t *testing.T) {
		for channel, want := range routing {
			assert.Equal(t, want, accountTransaction.RoutingChannel(channel).Type)
		}
	})
}

func TestCreateAccountTransactionRequestToAccountTransactionDTO(t *testing.T) {
	var (
		accountId                     = uuid.New()
		merchantId                    = uuid.New()
		transactionTimestamp          = time.Now()
		reasonType, reasonDescription = constant.ReasonTypeOtherReason, "Bla bla"
	)
	testcases := []struct {
		Name              string
		Request           *CreateAccountTransactionRequest
		Account           *account_model.Account
		Expected          *AccountTransaction
		ExpectedReference string
	}{
		{
			Name: "SUCCESS",
			Request: &CreateAccountTransactionRequest{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				ReasonType:           &reasonType,
				ReasonDescription:    &reasonDescription,
				Usecase:              constant.UseCaseDisbursement,
			},
			Account: &account_model.Account{
				UUID: accountId,
				Name: constant.TypeDisbursement,
			},
			Expected: &AccountTransaction{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				AccountID:            accountId,
				Reference:            constant.TypeDisbursement,
				ReasonType:           sql.NullString{Valid: true, String: reasonType},
				ReasonDescription:    sql.NullString{Valid: true, String: reasonDescription},
			},
			ExpectedReference: constant.TypeDisbursement,
		},
		{
			Name: "SUCCESS: DISBURSEMENT",
			Request: &CreateAccountTransactionRequest{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				ReasonType:           &reasonType,
				ReasonDescription:    &reasonDescription,
				Usecase:              constant.UseCaseDisbursement,
			},
			Account: &account_model.Account{
				UUID: accountId,
				Name: constant.TypeDisbursement,
			},
			Expected: &AccountTransaction{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				AccountID:            accountId,
				Reference:            constant.TypeDisbursement,
				ReasonType:           sql.NullString{Valid: true, String: reasonType},
				ReasonDescription:    sql.NullString{Valid: true, String: reasonDescription},
			},
			ExpectedReference: constant.TypeDisbursement,
		},
		{
			Name: "SUCCESS: DISBURSEMENT with undefined usecase",
			Request: &CreateAccountTransactionRequest{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				ReasonType:           &reasonType,
				ReasonDescription:    &reasonDescription,
			},
			Account: &account_model.Account{
				UUID: accountId,
				Name: constant.TypeDisbursement,
			},
			Expected: &AccountTransaction{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				AccountID:            accountId,
				Reference:            constant.TypeDisbursement,
				ReasonType:           sql.NullString{Valid: true, String: reasonType},
				ReasonDescription:    sql.NullString{Valid: true, String: reasonDescription},
			},
			ExpectedReference: constant.TypeDisbursement,
		},
		{
			Name: "SUCCESS: PAYMENT",
			Request: &CreateAccountTransactionRequest{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				ReasonType:           &reasonType,
				ReasonDescription:    &reasonDescription,
				Usecase:              constant.TypePayment,
			},
			Account: &account_model.Account{
				UUID: accountId,
				Name: constant.TypeDisbursement,
			},
			Expected: &AccountTransaction{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypeDisbursement,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				AccountID:            accountId,
				Reference:            constant.TypePayment,
				ReasonType:           sql.NullString{Valid: true, String: reasonType},
				ReasonDescription:    sql.NullString{Valid: true, String: reasonDescription},
			},
			ExpectedReference: constant.TypePayment,
		},
		{
			Name: "SUCCESS: PAYMENT with undefined usecase",
			Request: &CreateAccountTransactionRequest{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypePayment,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				ReasonType:           &reasonType,
				ReasonDescription:    &reasonDescription,
			},
			Account: &account_model.Account{
				UUID: accountId,
				Name: constant.TypeDisbursement,
			},
			Expected: &AccountTransaction{
				UUID:                 uuid.Nil,
				ReferenceID:          "REF-0001",
				MerchantID:           merchantId,
				Currency:             constant.CurrencyIDR,
				Credit:               100,
				Debit:                100,
				Type:                 constant.TypePayment,
				Channel:              constant.ChannelBankTransfer,
				Status:               constant.StatusSuccess,
				Remarks:              "remarks",
				TransactionTimestamp: transactionTimestamp,
				AdditionalInfo:       types.NullJSONText{},
				AccountID:            accountId,
				Reference:            constant.TypeDisbursement,
				ReasonType:           sql.NullString{Valid: true, String: reasonType},
				ReasonDescription:    sql.NullString{Valid: true, String: reasonDescription},
			},
			ExpectedReference: constant.TypeDisbursement,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			tcDTO := tc.Request.ToAccountTransactionDTO(tc.Account)

			if tc.Request.UUID == uuid.Nil {
				assert.NotNil(t, tcDTO.UUID)
			} else {
				assert.Equal(t, tcDTO.UUID, tc.Expected.UUID)
			}
			assert.Equal(t, tcDTO.ReferenceID, tc.Expected.ReferenceID)
			assert.Equal(t, tcDTO.MerchantID, tc.Expected.MerchantID)
			assert.Equal(t, tcDTO.Currency, tc.Expected.Currency)
			assert.Equal(t, tcDTO.Credit, tc.Expected.Credit)
			assert.Equal(t, tcDTO.Debit, tc.Expected.Debit)
			assert.Equal(t, tcDTO.Type, tc.Expected.Type)
			assert.Equal(t, tcDTO.Channel, tc.Expected.Channel)
			assert.Equal(t, tcDTO.Status, tc.Expected.Status)
			assert.Equal(t, tcDTO.Remarks, tc.Expected.Remarks)
			assert.Equal(t, tcDTO.TransactionTimestamp, tc.Expected.TransactionTimestamp)
			assert.Equal(t, tcDTO.AdditionalInfo, tc.Expected.AdditionalInfo)
			assert.Equal(t, tc.Account.UUID, tc.Expected.AccountID)
			assert.Equal(t, tc.ExpectedReference, tc.Expected.Reference)
			assert.Equal(t, reasonType, tc.Expected.ReasonType.String)
			assert.Equal(t, reasonDescription, tc.Expected.ReasonDescription.String)
		})
	}
}
