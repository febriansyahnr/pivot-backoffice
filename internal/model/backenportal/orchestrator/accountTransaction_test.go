package orchestrator_model

import (
	"testing"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	"github.com/stretchr/testify/assert"
)

func TestNewAccountTransaction(t *testing.T) {

	testCases := []struct {
		Name    string
		Input   *CreateNewTransactionRequest
		WantErr bool
	}{
		{
			Name:    "SUCCESS: Create transactions",
			WantErr: false,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            constant.ReferencePayment,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Now(),
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
				AdditionalInfo:       nil,
			},
		},
		{
			Name:    "SUCCESS: Create transactions with additional info",
			WantErr: false,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            constant.ReferencePayment,
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Now(),
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
				AdditionalInfo:       map[string]interface{}{"key": "value"},
			},
		},
		{
			Name:    "ERROR: Negative credit/debit",
			WantErr: true,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(-1000),
				Credit:               float64(-1000),
				Reference:            "unknown",
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Now(),
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
			},
		},
		{
			Name:    "ERROR: Incorrect Usecase",
			WantErr: true,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            "unknown",
				Type:                 constant.TypePayment,
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Now(),
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
			},
		},
		{
			Name:    "ERROR: Status incorrect",
			WantErr: true,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            constant.ReferencePayment,
				Type:                 constant.TypePayment,
				Status:               "unknown",
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Now(),
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
			},
		},
		{
			Name:    "ERROR: Empty timestamp",
			WantErr: true,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            constant.ReferencePayment,
				Type:                 constant.TypePayment + "S",
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Time{},
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
			},
		},
		{
			Name:    "ERROR: Invalid Additional Info",
			WantErr: true,
			Input: &CreateNewTransactionRequest{
				ReferenceID:          "123",
				Debit:                float64(1000),
				Reference:            constant.ReferencePayment,
				Type:                 constant.TypePayment + "S",
				Status:               constant.StatusSuccess,
				Channel:              constant.ChannelBalance,
				TransactionTimestamp: time.Time{},
				MerchantID:           uuid.New(),
				AccountID:            uuid.New(),
				Currency:             constant.CurrencyIDR,
				AdditionalInfo:       "invalid",
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			trx, err := NewAccountTransaction(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
				assert.NotNil(t, trx.UUID)
				assert.Equal(t, tc.Input.ReferenceID, trx.ReferenceID)
				assert.Equal(t, tc.Input.AccountID, trx.AccountID)
				assert.Equal(t, tc.Input.Currency, trx.Currency)
				assert.Equal(t, tc.Input.Credit, trx.Credit)
				assert.Equal(t, tc.Input.Debit, trx.Debit)
				assert.Equal(t, tc.Input.Type, trx.Type)
				assert.Equal(t, tc.Input.Channel, trx.Channel)
				assert.Equal(t, tc.Input.Status, trx.Status)
				assert.Equal(t, tc.Input.Remarks, trx.Remarks)
				assert.Equal(t, tc.Input.TransactionTimestamp, trx.TransactionTimestamp)
				assert.Equal(t, tc.Input.Reference, trx.Reference)
				assert.Equal(t, tc.Input.ReasonType, trx.ReasonType.String)
				assert.Equal(t, tc.Input.ReasonDescription, trx.ReasonDescription.String)
				assert.NotEmpty(t, trx.CreatedAt)
				assert.NotEmpty(t, trx.UpdatedAt)

			}
		})

	}
}

func TestValidateUsecase(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   string
		WantErr bool
	}{
		{
			Name:    "SUCCESS: Payment",
			Input:   constant.ReferencePayment,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Disbursement",
			Input:   constant.ReferenceDisbursement,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Wallet",
			Input:   constant.ReferenceWallet,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Platform",
			Input:   constant.ReferencePlatform,
			WantErr: false,
		},
		{
			Name:    "ERROR: Unknown",
			Input:   "unknown",
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := ValidateUseCase(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestValidateType(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   string
		WantErr bool
	}{
		{
			Name:    "SUCCESS: Empty Type",
			Input:   "",
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Payment",
			Input:   constant.TypePayment,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Disbursement",
			Input:   constant.TypeDisbursement,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type TopUp",
			Input:   constant.TypeTopUp,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type ManualAdjust",
			Input:   constant.TypeManualAdjust,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Fee",
			Input:   constant.TypeFee,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type AccountInquiryFee",
			Input:   constant.TypeAccountInquiryFee,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Wallet TopUp",
			Input:   constant.TypeWalletTopUp,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Wallet Transfer",
			Input:   constant.TypeWalletTransfer,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Wallet Withdrawal",
			Input:   constant.TypeWalletWithdrawal,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Type Wallet Bill Payment",
			Input:   constant.TypeWalletBillPayment,
			WantErr: false,
		},
		{
			Name:    "ERROR: Unknown Type",
			Input:   "unknown",
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateType(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestValidateChannel(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   string
		WantErr bool
	}{
		{
			Name:    "SUCCESS: Empty Channel",
			Input:   "",
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Balance",
			Input:   constant.ChannelBalance,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Virtual Account",
			Input:   constant.ChannelVirtualAccount,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Credit Card",
			Input:   constant.ChannelCreditCard,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Bank Transfer",
			Input:   constant.ChannelBankTransfer,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Manual Transfer",
			Input:   constant.ChannelManualTransfer,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Balance Adjustment",
			Input:   constant.ChannelBalanceAdjustment,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: PPOB Channel",
			Input:   constant.ChannelPPOB,
			WantErr: false,
		},
		{
			Name:    "ERROR: Unknown channel",
			Input:   "unknown",
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateChannel(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}

}

func TestValidateStatus(t *testing.T) {

	testCases := []struct {
		Name    string
		Input   string
		WantErr bool
	}{
		{
			Name:    "SUCCESS: Success status",
			Input:   constant.StatusSuccess,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Failed status",
			Input:   constant.StatusFailed,
			WantErr: false,
		},
		{
			Name:    "SUCCESS: Pending status",
			Input:   constant.StatusPending,
			WantErr: false,
		},
		{
			Name:    "ERROR: Unknown status",
			Input:   "",
			WantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := validateStatus(tc.Input)
			if tc.WantErr {
				if !assert.NotNil(t, err) {
					t.Errorf("ValidateStatus() error = %v, wantErr %v", err, tc.WantErr)
				}
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

// TestNewAccountTransactionUUIDError tests the error handling of NewAccountTransaction when UUID generation fails
func TestNewAccountTransactionUUIDError(t *testing.T) {
	// Save the original generator and restore it after the test
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Replace the UUID generator with one that returns an error
	defaultUUIDGenerator = func() (uuid.UUID, error) {
		return uuid.Nil, errors.New("simulated UUID generation error")
	}

	// Create a valid request
	request := &CreateNewTransactionRequest{
		ReferenceID:          "123",
		Debit:                float64(1000),
		Reference:            constant.ReferencePayment,
		Type:                 constant.TypePayment,
		Status:               constant.StatusSuccess,
		Channel:              constant.ChannelBalance,
		TransactionTimestamp: time.Now(),
		MerchantID:           uuid.New(),
		AccountID:            uuid.New(),
		Currency:             constant.CurrencyIDR,
	}

	// Call the function with the mock generator
	result, err := NewAccountTransaction(request)

	// Verify that the error is propagated and the result is nil
	assert.Error(t, err, "Function should return an error when UUID generation fails")
	assert.Nil(t, result, "Result should be nil when an error occurs")
	assert.Contains(t, err.Error(), "simulated UUID generation error", "Error message should be propagated")
}

func TestGetCreditcardMetadataFromAdditionalInfo(t *testing.T) {
	// Test data
	validJsonData := `{
		"expiredAt": "2025-04-30T15:04:05Z",
		"feeDetail": {
			"type": "PAYMENT",
			"amount": 2000,
			"method": "CREDIT_CARD",
			"taxType": "NON_PKP",
			"taxAmount": 0,
			"amountType": "AMOUNT_PERCENTAGE",
			"percentage": 2.5,
			"finalAmount": 4500,
			"deductionType": "DIRECT",
			"referenceType": "",
			"taxPercentage": 0
		},
		"chargeStatus": "SUCCESS",
		"methodDetail": {
			"card": {
				"last4": "0010",
				"first6": "444000",
				"first8": "44400000",
				"expYear": 0,
				"expMonth": 0,
				"fingerprint": "0077187d-f69d-4b2c-a9f9-99aeb6919dda",
				"cardBrand": "VISA",
				"countryCode": "ID",
				"binInformations": {
					"type": "DEBIT",
					"brand": "VISA",
					"country": "ID",
					"issuingBank": "BRI_S2I"
				},
				"authorizationResult": {
					"stan": "240399",
					"avsResult": "",
					"cvvResult": "",
					"authorizedAmount": {
						"value": 100000,
						"currency": "IDR"
					},
					"acquirerReferenceNumber": "123456789012345",
					"issuerAuthorizationCode": "00",
					"retrievalReferenceNumber": "TRXCC8fb53668d1c717458198311"
				},
				"authenticationResult": {
					"eciCode": "05",
					"threeDsMethod": "",
					"threeDsResult": "AUTHENTICATION_SUCCESSFUL",
					"threeDsVersion": "2.2.0"
				}
			}
		},
		"reconReferenceNo": "PAYCC7e47d09ca94c1745819846",
		"settlementDetail": {
			"type": "T+5",
			"dayType": "ANYDAY",
			"endCutOffTime": "",
			"executionTime": "",
			"startCutOffTime": ""
		},
		"processorTransactionId": ""
	}`

	emptyJsonData := `{}`
	invalidJsonData := `{"invalid json`

	tests := []struct {
		name           string
		additionalInfo types.NullJSONText
		wantErr        bool
		validateFunc   func(t *testing.T, metadata *creditcardModel.CreditcardMetadata)
	}{
		{
			name: "Success with valid JSON",
			additionalInfo: types.NullJSONText{
				JSONText: []byte(validJsonData),
				Valid:    true,
			},
			wantErr: false,
			validateFunc: func(t *testing.T, metadata *creditcardModel.CreditcardMetadata) {
				assert.NotNil(t, metadata)
				assert.Equal(t, "SUCCESS", metadata.ProcessorStatus)
				assert.NotNil(t, metadata.CardData)
				assert.Equal(t, "0010", metadata.CardData.Last4Digit)
				assert.Equal(t, "44400000", metadata.CardData.First8Digit)
				assert.Equal(t, "0077187d-f69d-4b2c-a9f9-99aeb6919dda", metadata.CardData.Fingerprint)
				assert.Equal(t, "DEBIT", metadata.CardData.CardType)
				assert.Equal(t, "VISA", metadata.CardData.CardBrand)
				assert.Equal(t, "BRI_S2I", metadata.CardData.CardIssuing)
				assert.Equal(t, "ID", metadata.CardData.CountryCode)

				assert.NotNil(t, metadata.AuthorizationData)
				assert.NotNil(t, metadata.AuthenticationData)
				assert.NotNil(t, metadata.FeeDetail)
				assert.Equal(t, "PAYMENT", metadata.FeeDetail.Type)
			},
		},
		{
			name: "Success with empty JSON",
			additionalInfo: types.NullJSONText{
				JSONText: []byte(emptyJsonData),
				Valid:    true,
			},
			wantErr: true,
			validateFunc: func(t *testing.T, metadata *creditcardModel.CreditcardMetadata) {
				assert.Nil(t, metadata)
			},
		},
		{
			name: "Error with invalid JSON",
			additionalInfo: types.NullJSONText{
				JSONText: []byte(invalidJsonData),
				Valid:    true,
			},
			wantErr: true,
			validateFunc: func(t *testing.T, metadata *creditcardModel.CreditcardMetadata) {
				assert.Nil(t, metadata)
			},
		},
		{
			name: "Error with empty AdditionalInfo",
			additionalInfo: types.NullJSONText{
				JSONText: []byte{},
				Valid:    false,
			},
			wantErr: true,
			validateFunc: func(t *testing.T, metadata *creditcardModel.CreditcardMetadata) {
				assert.Nil(t, metadata)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trx := &AccountTransactionWithUseCase{
				AdditionalInfo: tt.additionalInfo,
			}

			metadata, err := trx.GetCreditcardMetadataFromAdditionalInfo()

			if tt.wantErr {
				assert.Error(t, err)
				tt.validateFunc(t, metadata)
			} else {
				assert.NoError(t, err)
				tt.validateFunc(t, metadata)
			}
		})
	}
}
