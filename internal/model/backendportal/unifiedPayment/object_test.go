package unifiedPaymentModel

import (
	"database/sql"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChargePaymentMethodDetailQrScan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    ChargePaymentMethodDetailQr
		wantErr bool
	}{
		{
			name:  "valid json",
			value: []byte(`{"acquirer":"DANA","qrContent":"qr-content-123","qrUrl":"https://example.com/qr","retrievalReferenceNumber":"ref-123","issuerName":"Test Bank","expiryAt":"2023-01-02T15:04:05Z","merchantName":"Test Merchant"}`),
			want: ChargePaymentMethodDetailQr{
				Acquirer:                 "DANA",
				QrContent:                "qr-content-123",
				QrUrl:                    "https://example.com/qr",
				RetrievalReferenceNumber: "ref-123",
				IssuerName:               "Test Bank",
				ExpiryAt:                 time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
				MerchantName:             "Test Merchant",
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{"acquirer":123}`),
			want:    ChargePaymentMethodDetailQr{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    ChargePaymentMethodDetailQr{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qr := &ChargePaymentMethodDetailQr{}
			err := qr.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChargePaymentMethodDetailQr.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(*qr, tt.want) {
				t.Errorf("ChargePaymentMethodDetailQr.Scan() = %v, want %v", *qr, tt.want)
			}
		})
	}
}

func TestChargePaymentMethodDetailVirtualAccountScan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    ChargePaymentMethodDetailVirtualAccount
		wantErr bool
	}{
		{
			name:  "valid json",
			value: []byte(`{"channel":"BCA","virtualAccountNumber":"1234567890","virtualAccountName":"John Doe","expiryAt":"2023-01-02T15:04:05Z"}`),
			want: ChargePaymentMethodDetailVirtualAccount{
				Channel:              "BCA",
				VirtualAccountNumber: "1234567890",
				VirtualAccountName:   "John Doe",
				ExpiryAt:             time.Date(2023, 1, 2, 15, 4, 5, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{"channel":123}`),
			want:    ChargePaymentMethodDetailVirtualAccount{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    ChargePaymentMethodDetailVirtualAccount{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			va := &ChargePaymentMethodDetailVirtualAccount{}
			err := va.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChargePaymentMethodDetailVirtualAccount.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(*va, tt.want) {
				t.Errorf("ChargePaymentMethodDetailVirtualAccount.Scan() = %v, want %v", *va, tt.want)
			}
		})
	}
}

func TestChargePaymentMethodDetailCardScan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    ChargePaymentMethodDetailCard
		wantErr bool
	}{
		{
			name: "valid json",
			value: []byte(`{
				"first6":"123456",
				"first8":"12345678",
				"last4":"9012",
				"expMonth":"12",
				"expYear":"25",
				"fingerprint":"fp123",
				"binInformations":{
					"type":"CREDIT",
					"issuingBank":"Test Bank",
					"brand":"VISA",
					"country":"ID"
				},
				"authenticationResult":{},
				"authorizationResult":{}
			}`),
			want: ChargePaymentMethodDetailCard{
				First6:      "123456",
				First8:      "12345678",
				Last4:       "9012",
				ExpMonth:    "12",
				ExpYear:     "25",
				Fingerprint: "fp123",
				BinInformations: ChargePaymentMethodDetailBinInformation{
					Type:        "CREDIT",
					IssuingBank: "Test Bank",
					Brand:       "VISA",
					Country:     "ID",
				},
				AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{},
				AuthorizationResult:  &ChargePaymentMethodDetailCardAuthorizationResult{},
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{"first6":123}`),
			want:    ChargePaymentMethodDetailCard{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    ChargePaymentMethodDetailCard{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &ChargePaymentMethodDetailCard{}
			err := cc.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("ChargePaymentMethodDetailCard.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(*cc, tt.want) {
				t.Errorf("ChargePaymentMethodDetailCard.Scan() = %v, want %v", *cc, tt.want)
			}
		})
	}
}
func TestAmountScan(t *testing.T) {
	tests := []struct {
		name    string
		value   interface{}
		want    Amount
		wantErr bool
	}{
		{
			name:  "valid json",
			value: []byte(`{"value":10000.50,"currency":"IDR"}`),
			want: Amount{
				Value:    10000.50,
				Currency: "IDR",
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			value:   []byte(`{"value":"not-a-number"}`),
			want:    Amount{},
			wantErr: true,
		},
		{
			name:    "nil value",
			value:   nil,
			want:    Amount{},
			wantErr: false,
		},
		{
			name:  "zero value",
			value: []byte(`{"value":0,"currency":"USD"}`),
			want: Amount{
				Value:    0,
				Currency: "USD",
			},
			wantErr: false,
		},
		{
			name:  "negative value",
			value: []byte(`{"value":-100.25,"currency":"EUR"}`),
			want: Amount{
				Value:    -100.25,
				Currency: "EUR",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Amount{}
			err := a.Scan(tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("Amount.Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(*a, tt.want) {
				t.Errorf("Amount.Scan() = %v, want %v", *a, tt.want)
			}
		})
	}
}

func TestStructThatHaveScanFunction(t *testing.T) {
	tests := []struct {
		src        string
		dst        sql.Scanner
		wantResult any
	}{
		{
			src: `{"channel": "PIVOTPAY", "referenceNo": "REF001", "webRedirectUrl": "https://localhost", "partnerReferenceNo": "PARTNER-REFF001"}`,
			dst: &ChargePaymentMethodDetailEwallet{},
			wantResult: &ChargePaymentMethodDetailEwallet{
				WebRedirectURL:     "https://localhost", // NOSONAR
				ReferenceNo:        "REF001",            // NOSONAR
				PartnerReferenceNo: "PARTNER-REFF001",   // NOSONAR
				Channel:            "PIVOTPAY",          // NOSONAR
			},
		},
	}
	for _, test := range tests {
		require.NotNil(t, test.dst)

		assert.NoError(t, test.dst.Scan([]byte(test.src)))
		assert.Equal(t, test.wantResult, test.dst)
	}
}

func TestPaymentMethodOptions_Merge(t *testing.T) {
	expiryTime := time.Now()

	tests := []struct {
		name     string
		override *PaymentMethodOptions
		base     *PaymentMethodOptions
		want     *PaymentMethodOptions
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
			},
			want: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
			},
		},
		{
			name: "base nil should return override",
			override: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
			},
			base: nil,
			want: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
			},
		},
		{
			name: "override takes precedence for VirtualAccount",
			override: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
			},
			base: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
			},
			want: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptions{
				Card: &PaymentMethodOptionCard{CaptureMethod: "manual"},
			},
			base: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
				QR:             &PaymentMethodOptionQR{ExpiryAt: &expiryTime},
			},
			want: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
				QR:             &PaymentMethodOptionQR{ExpiryAt: &expiryTime},
				Card:           &PaymentMethodOptionCard{CaptureMethod: "manual"},
			},
		},
		{
			name: "all payment methods merged correctly",
			override: &PaymentMethodOptions{
				Card:    &PaymentMethodOptionCard{CaptureMethod: "automatic"},
				Ewallet: &PaymentMethodOptionEwallet{Channel: "DANA"},
			},
			base: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
				QR:             &PaymentMethodOptionQR{ExpiryAt: &expiryTime},
				Card:           &PaymentMethodOptionCard{ThreeDsMethod: "NEVER"},
			},
			want: &PaymentMethodOptions{
				VirtualAccount: &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
				QR:             &PaymentMethodOptionQR{ExpiryAt: &expiryTime},
				Card:           &PaymentMethodOptionCard{CaptureMethod: "automatic", ThreeDsMethod: "NEVER"},
				Ewallet:        &PaymentMethodOptionEwallet{Channel: "DANA"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.Merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionVirtualAccount_merge(t *testing.T) {
	expiryTime1 := time.Now()
	expiryTime2 := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		override *PaymentMethodOptionVirtualAccount
		base     *PaymentMethodOptionVirtualAccount
		want     *PaymentMethodOptionVirtualAccount
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base:     &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
			want:     &PaymentMethodOptionVirtualAccount{Channel: "BCA"},
		},
		{
			name:     "base nil should return override",
			override: &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
			base:     nil,
			want:     &PaymentMethodOptionVirtualAccount{Channel: "MANDIRI"},
		},
		{
			name: "override takes precedence for all fields",
			override: &PaymentMethodOptionVirtualAccount{
				Channel:               "MANDIRI",
				VirtualAccountTrxType: "OPEN",
				VirtualAccountName:    "Override Name",
				VirtualAccountNumber:  "9876543210",
				ExpiryAt:              &expiryTime1,
			},
			base: &PaymentMethodOptionVirtualAccount{
				Channel:               "BCA",
				VirtualAccountTrxType: "CLOSE",
				VirtualAccountName:    "Base Name",
				VirtualAccountNumber:  "1234567890",
				ExpiryAt:              &expiryTime2,
			},
			want: &PaymentMethodOptionVirtualAccount{
				Channel:               "MANDIRI",
				VirtualAccountTrxType: "OPEN",
				VirtualAccountName:    "Override Name",
				VirtualAccountNumber:  "9876543210",
				ExpiryAt:              &expiryTime1,
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptionVirtualAccount{
				Channel: "MANDIRI",
			},
			base: &PaymentMethodOptionVirtualAccount{
				Channel:               "BCA",
				VirtualAccountTrxType: "CLOSE",
				VirtualAccountName:    "Base Name",
				VirtualAccountNumber:  "1234567890",
				ExpiryAt:              &expiryTime2,
			},
			want: &PaymentMethodOptionVirtualAccount{
				Channel:               "MANDIRI",
				VirtualAccountTrxType: "CLOSE",
				VirtualAccountName:    "Base Name",
				VirtualAccountNumber:  "1234567890",
				ExpiryAt:              &expiryTime2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionQR_merge(t *testing.T) {
	expiryTime1 := time.Now()
	expiryTime2 := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		override *PaymentMethodOptionQR
		base     *PaymentMethodOptionQR
		want     *PaymentMethodOptionQR
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime1},
			want:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime1},
		},
		{
			name:     "base nil should return override",
			override: &PaymentMethodOptionQR{ExpiryAt: &expiryTime2},
			base:     nil,
			want:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime2},
		},
		{
			name:     "override takes precedence",
			override: &PaymentMethodOptionQR{ExpiryAt: &expiryTime1},
			base:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime2},
			want:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime1},
		},
		{
			name:     "base fills missing expiry in override",
			override: &PaymentMethodOptionQR{},
			base:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime2},
			want:     &PaymentMethodOptionQR{ExpiryAt: &expiryTime2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionCard_merge(t *testing.T) {
	expiryTime1 := time.Now()
	expiryTime2 := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		override *PaymentMethodOptionCard
		base     *PaymentMethodOptionCard
		want     *PaymentMethodOptionCard
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
				ThreeDsMethod: "AUTOMATIC",
			},
			want: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
				ThreeDsMethod: "AUTOMATIC",
			},
		},
		{
			name: "base nil should return override",
			override: &PaymentMethodOptionCard{
				CaptureMethod: "automatic",
				ThreeDsMethod: "NEVER",
			},
			base: nil,
			want: &PaymentMethodOptionCard{
				CaptureMethod: "automatic",
				ThreeDsMethod: "NEVER",
			},
		},
		{
			name: "override takes precedence for all fields",
			override: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
				ThreeDsMethod: "EXTERNAL",
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID123",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: true,
				},
				ExpiryAt: &expiryTime1,
			},
			base: &PaymentMethodOptionCard{
				CaptureMethod: "automatic",
				ThreeDsMethod: "AUTOMATIC",
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID456",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: false,
				},
				ExpiryAt: &expiryTime2,
			},
			want: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
				ThreeDsMethod: "EXTERNAL",
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID123",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: true,
				},
				ExpiryAt: &expiryTime1,
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
			},
			base: &PaymentMethodOptionCard{
				CaptureMethod: "automatic",
				ThreeDsMethod: "AUTOMATIC",
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID456",
				},
				ExpiryAt: &expiryTime2,
			},
			want: &PaymentMethodOptionCard{
				CaptureMethod: "manual",
				ThreeDsMethod: "AUTOMATIC",
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID456",
				},
				ExpiryAt: &expiryTime2,
			},
		},
		{
			name: "nested processing config and installment are merged",
			override: &PaymentMethodOptionCard{
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID123",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: true,
				},
			},
			base: &PaymentMethodOptionCard{
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID456",
					MerchantIdTag:  "TAG123",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: false,
					Plan:    map[string]interface{}{"tenor": 3},
				},
			},
			want: &PaymentMethodOptionCard{
				ProcessingConfig: &PaymentMethodOptionCardProcessingConfig{
					BankMerchantId: "MID123",
					MerchantIdTag:  "TAG123",
				},
				Installment: &PaymentMethodOptionCardInstallment{
					Enabled: true,
					Plan:    map[string]interface{}{"tenor": 3},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionCardProcessingConfig_merge(t *testing.T) {
	tests := []struct {
		name     string
		override *PaymentMethodOptionCardProcessingConfig
		base     *PaymentMethodOptionCardProcessingConfig
		want     *PaymentMethodOptionCardProcessingConfig
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID123",
				MerchantIdTag:  "TAG123",
			},
			want: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID123",
				MerchantIdTag:  "TAG123",
			},
		},
		{
			name: "base nil should return override",
			override: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
				MerchantIdTag:  "TAG456",
			},
			base: nil,
			want: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
				MerchantIdTag:  "TAG456",
			},
		},
		{
			name: "override takes precedence for all fields",
			override: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
				MerchantIdTag:  "TAG456",
			},
			base: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID123",
				MerchantIdTag:  "TAG123",
			},
			want: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
				MerchantIdTag:  "TAG456",
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
			},
			base: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID123",
				MerchantIdTag:  "TAG123",
			},
			want: &PaymentMethodOptionCardProcessingConfig{
				BankMerchantId: "MID456",
				MerchantIdTag:  "TAG123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionCardInstallment_merge(t *testing.T) {
	plan1 := map[string]interface{}{"tenor": 3, "rate": 0.5}
	plan2 := map[string]interface{}{"tenor": 6, "rate": 1.0}
	availablePlans1 := []interface{}{plan1}
	availablePlans2 := []interface{}{plan2}

	tests := []struct {
		name     string
		override *PaymentMethodOptionCardInstallment
		base     *PaymentMethodOptionCardInstallment
		want     *PaymentMethodOptionCardInstallment
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
			want: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
		},
		{
			name: "base nil should return override",
			override: &PaymentMethodOptionCardInstallment{
				Enabled:        false,
				AvailablePlans: availablePlans2,
				Plan:           plan2,
			},
			base: nil,
			want: &PaymentMethodOptionCardInstallment{
				Enabled:        false,
				AvailablePlans: availablePlans2,
				Plan:           plan2,
			},
		},
		{
			name: "override enabled false takes precedence over base enabled true",
			override: &PaymentMethodOptionCardInstallment{
				Enabled: false,
			},
			base: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
			want: &PaymentMethodOptionCardInstallment{
				Enabled:        false,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
		},
		{
			name: "override takes precedence for all fields",
			override: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans2,
				Plan:           plan2,
			},
			base: &PaymentMethodOptionCardInstallment{
				Enabled:        false,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
			want: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans2,
				Plan:           plan2,
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptionCardInstallment{
				Enabled: true,
			},
			base: &PaymentMethodOptionCardInstallment{
				Enabled:        false,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
			want: &PaymentMethodOptionCardInstallment{
				Enabled:        true,
				AvailablePlans: availablePlans1,
				Plan:           plan1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPaymentMethodOptionEwallet_merge(t *testing.T) {
	expiryTime1 := time.Now()
	expiryTime2 := time.Now().Add(time.Hour)

	tests := []struct {
		name     string
		override *PaymentMethodOptionEwallet
		base     *PaymentMethodOptionEwallet
		want     *PaymentMethodOptionEwallet
	}{
		{
			name:     "both nil should return nil",
			override: nil,
			base:     nil,
			want:     nil,
		},
		{
			name:     "override nil should return base",
			override: nil,
			base: &PaymentMethodOptionEwallet{
				Channel:  "SHOPEEPAY",
				ExpiryAt: &expiryTime1,
			},
			want: &PaymentMethodOptionEwallet{
				Channel:  "SHOPEEPAY",
				ExpiryAt: &expiryTime1,
			},
		},
		{
			name: "base nil should return override",
			override: &PaymentMethodOptionEwallet{
				Channel:  "DANA",
				ExpiryAt: &expiryTime2,
			},
			base: nil,
			want: &PaymentMethodOptionEwallet{
				Channel:  "DANA",
				ExpiryAt: &expiryTime2,
			},
		},
		{
			name: "override takes precedence for all fields",
			override: &PaymentMethodOptionEwallet{
				Channel:  "DANA",
				ExpiryAt: &expiryTime1,
			},
			base: &PaymentMethodOptionEwallet{
				Channel:  "SHOPEEPAY",
				ExpiryAt: &expiryTime2,
			},
			want: &PaymentMethodOptionEwallet{
				Channel:  "DANA",
				ExpiryAt: &expiryTime1,
			},
		},
		{
			name: "base fills missing fields in override",
			override: &PaymentMethodOptionEwallet{
				Channel: "DANA",
			},
			base: &PaymentMethodOptionEwallet{
				Channel:  "SHOPEEPAY",
				ExpiryAt: &expiryTime2,
			},
			want: &PaymentMethodOptionEwallet{
				Channel:  "DANA",
				ExpiryAt: &expiryTime2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.override.merge(tt.base)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChargePaymentMethodDetails_GetCardMID(t *testing.T) {
	tests := []struct {
		name     string
		details  ChargePaymentMethodDetails
		expected string
	}{
		{
			name:     "returns empty string when Card is nil",
			details:  ChargePaymentMethodDetails{},
			expected: "",
		},
		{
			name: "returns empty string when MIDInfo is nil",
			details: ChargePaymentMethodDetails{
				Card: &ChargePaymentMethodDetailCard{},
			},
			expected: "",
		},
		{
			name: "returns MID when both Card and MIDInfo are present",
			details: ChargePaymentMethodDetails{
				Card: &ChargePaymentMethodDetailCard{
					MIDInfo: &MIDInfo{
						MID:      "MID123456",
						Acquirer: "BANK_A",
						Type:     "CREDIT",
					},
				},
			},
			expected: "MID123456",
		},
		{
			name: "returns empty string when MID is empty",
			details: ChargePaymentMethodDetails{
				Card: &ChargePaymentMethodDetailCard{
					MIDInfo: &MIDInfo{
						MID:      "",
						Acquirer: "BANK_A",
						Type:     "CREDIT",
					},
				},
			},
			expected: "",
		},
		{
			name: "returns MID correctly with other Card fields populated",
			details: ChargePaymentMethodDetails{
				Card: &ChargePaymentMethodDetailCard{
					First6:         "411111",
					Last4:          "1234",
					Fingerprint:    "fp_abc123",
					CardHolderName: "John Doe",
					MIDInfo: &MIDInfo{
						MID:      "MID789012",
						Acquirer: "BANK_B",
						Type:     "DEBIT",
					},
				},
			},
			expected: "MID789012",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.details.GetCardMID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNaturalPaymentFailureMessage(t *testing.T) {
	tests := []struct {
		name              string
		paymentMethod     string
		failureCode       string
		authorizationCode string
		expected          string
	}{
		// DECLINED_BY_CHANNEL - Card with specific authorization codes
		{
			name:              "card declined - expired card (code 54)",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "54",
			expected:          "Your card has expired. Please use another card or payment method.",
		},
		{
			name:              "card declined - expired card (code 101)",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "101",
			expected:          "Your card has expired. Please use another card or payment method.",
		},
		{
			name:              "card declined - invalid CVV",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "N7",
			expected:          "Card verification failed. Please check your card details and try again.",
		},
		{
			name:              "card declined - generic",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "05",
			expected:          "Payment was declined by issuer. Please contact your card issuer or use another payment method.",
		},
		{
			name:              "card declined - no authorization code",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "",
			expected:          "Payment was declined by issuer. Please contact your card issuer or use another payment method.",
		},
		{
			name:              "ewallet declined",
			paymentMethod:     "EWALLET",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "",
			expected:          "Payment was declined by issuer. Please contact your e-wallet issuer or use another payment method.",
		},

		// INVALID_ACCOUNT
		{
			name:              "card invalid account",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "INVALID_ACCOUNT",
			authorizationCode: "",
			expected:          "Your card is invalid. Please use another card or payment method.",
		},
		{
			name:              "ewallet invalid account",
			paymentMethod:     "EWALLET",
			failureCode:       "INVALID_ACCOUNT",
			authorizationCode: "",
			expected:          "Your account is not valid. Please use another payment method.",
		},

		// AUTHENTICATION_FAILED
		{
			name:              "authentication failed",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "AUTHENTICATION_FAILED",
			authorizationCode: "",
			expected:          "Payment could not be verified. Please contact your card issuer or use another card.",
		},

		// SUSPECTED_FRAUD
		{
			name:              "suspected fraud",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "SUSPECTED_FRAUD",
			authorizationCode: "",
			expected:          "Payment could not be completed. Please contact your card issuer.",
		},

		// BLOCKED_BY_FDS
		{
			name:              "blocked by FDS",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "BLOCKED_BY_FDS",
			authorizationCode: "",
			expected:          "Payment could not be completed at this time. Please use another card or payment method.",
		},

		// REQUIRE_REVIEW
		{
			name:              "require review",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "REQUIRE_REVIEW",
			authorizationCode: "",
			expected:          "Payment in process. Please wait a moment.",
		},

		// INSUFFICIENT_FUND
		{
			name:              "insufficient fund",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "INSUFFICIENT_FUND",
			authorizationCode: "",
			expected:          "Insufficient funds. Please add funds or use another payment method.",
		},

		// CHANNEL_UNAVAILABLE
		{
			name:              "channel unavailable",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "CHANNEL_UNAVAILABLE",
			authorizationCode: "",
			expected:          "Payment service is temporarily unavailable. Please try again later or use another payment method.",
		},

		// CANCELLED_BY_USER
		{
			name:              "cancelled by user",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "CANCELLED_BY_USER",
			authorizationCode: "",
			expected:          "Payment was cancelled.",
		},

		// CHARGE_EXPIRED
		{
			name:              "charge expired",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "CHARGE_EXPIRED",
			authorizationCode: "",
			expected:          "Payment session expired. Please try again.",
		},

		// EXCEEDED_CAPTURE_PERIOD
		{
			name:              "exceeded capture period",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "EXCEEDED_CAPTURE_PERIOD",
			authorizationCode: "",
			expected:          "Payment could not be completed. Please try again.",
		},

		// Unknown failure code
		{
			name:              "unknown failure code",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "UNKNOWN_ERROR",
			authorizationCode: "",
			expected:          "",
		},

		// Edge cases
		{
			name:              "empty payment method",
			paymentMethod:     "",
			failureCode:       "DECLINED_BY_CHANNEL",
			authorizationCode: "",
			expected:          "",
		},
		{
			name:              "empty failure code",
			paymentMethod:     "CREDIT_CARD",
			failureCode:       "",
			authorizationCode: "",
			expected:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ChargePaymentMethodDetails{
				Card: &ChargePaymentMethodDetailCard{
					AuthorizationResult: &ChargePaymentMethodDetailCardAuthorizationResult{
						IssuerAuthorizationCode: tt.authorizationCode,
					},
				},
			}
			result := s.GetNaturalPaymentFailureMessage(tt.paymentMethod, tt.failureCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}
