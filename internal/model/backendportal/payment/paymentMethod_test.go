package paymentModel

import (
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/backendportal/paymentMethod"

	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/assert"
)

func TestToPlatformResponseModel(t *testing.T) {
	description := "description"
	logo := "logo"
	bankName := "bankname"

	paymentMethod := &PaymentMethodWithPivot{
		PaymentMethod: PaymentMethod{
			UUID:        "uuid",
			Type:        "type",
			Category:    "category",
			Name:        "name",
			Description: &description,
			Logo:        &logo,
			Acquirer:    "acquirer",
			BankName:    &bankName,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		MerchantID: "merchantid",
		IsActive:   true,
	}

	response := paymentMethod.ToPlatformResponseModel()
	assert.Equal(t, response.UUID, paymentMethod.UUID)
	assert.Equal(t, response.Name, paymentMethod.Name)
	assert.Equal(t, response.Description, paymentMethod.Description)
	assert.Equal(t, response.Logo, paymentMethod.Logo)
	assert.Equal(t, response.Acquirer, paymentMethod.Acquirer)
	assert.Equal(t, response.BankName, paymentMethod.BankName)
}

func TestValidatePaymentMethod(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "SUCCESS: Virtual Account",
			input:   constant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
			wantErr: false,
		},
		{
			name:    "SUCCESS: Credit Card",
			input:   constant.PAYMENT_METHOD_CREDIT_CARD,
			wantErr: false,
		},
		{
			name:    "SUCCESS: QRIS",
			input:   constant.PAYMENT_METHOD_QRIS,
			wantErr: false,
		},
		{
			name:    "ERROR: Incorrect Payment",
			input:   "Stones",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePaymentMethod(tc.input)
			if tc.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestPaymentMethod_UnmarshalConfigObj(t *testing.T) {
	testCases := []struct {
		name           string
		paymentMethod  *PaymentMethod
		expectedConfig *PaymentMethodConfig
		expectedNil    bool
	}{
		{
			name: "SUCCESS: Valid config with expiry duration in minutes",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":30,"unit":"minutes"}}`),
					Valid:    true,
				},
			},
			expectedConfig: &PaymentMethodConfig{
				ExpiryConfig: PaymentMethodExpiryConfig{
					Duration: 30,
					Unit:     "minutes",
				},
			},
			expectedNil: false,
		},
		{
			name: "SUCCESS: Valid config with expiry duration in hours",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":24,"unit":"hours"}}`),
					Valid:    true,
				},
			},
			expectedConfig: &PaymentMethodConfig{
				ExpiryConfig: PaymentMethodExpiryConfig{
					Duration: 24,
					Unit:     "hours",
				},
			},
			expectedNil: false,
		},
		{
			name: "SUCCESS: Valid config with expiry duration in days",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":7,"unit":"days"}}`),
					Valid:    true,
				},
			},
			expectedConfig: &PaymentMethodConfig{
				ExpiryConfig: PaymentMethodExpiryConfig{
					Duration: 7,
					Unit:     "days",
				},
			},
			expectedNil: false,
		},
		{
			name: "ERROR: Invalid JSON in config",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{invalid json}`),
					Valid:    true,
				},
			},
			expectedConfig: nil,
			expectedNil:    true,
		},
		{
			name: "ERROR: Empty JSON object",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{}`),
					Valid:    true,
				},
			},
			expectedConfig: &PaymentMethodConfig{
				ExpiryConfig: PaymentMethodExpiryConfig{
					Duration: 0,
					Unit:     "",
				},
			},
			expectedNil: false,
		},
		{
			name: "ERROR: Invalid config - not valid NullJSONText",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":30,"unit":"minutes"}}`),
					Valid:    false,
				},
			},
			expectedConfig: nil,
			expectedNil:    true,
		},
		{
			name: "ERROR: Null config",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: nil,
					Valid:    false,
				},
			},
			expectedConfig: nil,
			expectedNil:    true,
		},
		{
			name: "SUCCESS: Valid config with zero duration",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":0,"unit":"seconds"}}`),
					Valid:    true,
				},
			},
			expectedConfig: &PaymentMethodConfig{
				ExpiryConfig: PaymentMethodExpiryConfig{
					Duration: 0,
					Unit:     "seconds",
				},
			},
			expectedNil: false,
		},
		{
			name: "ERROR: Malformed JSON - missing closing brace",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`{"expiryConfig":{"duration":30,"unit":"minutes"}`),
					Valid:    true,
				},
			},
			expectedConfig: nil,
			expectedNil:    true,
		},
		{
			name: "ERROR: Invalid JSON type - string instead of object",
			paymentMethod: &PaymentMethod{
				Config: types.NullJSONText{
					JSONText: []byte(`"invalid"`),
					Valid:    true,
				},
			},
			expectedConfig: nil,
			expectedNil:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.paymentMethod.UnmarshalConfigObj()

			if tc.expectedNil {
				assert.Nil(t, tc.paymentMethod.ConfigObj)
			} else {
				assert.NotNil(t, tc.paymentMethod.ConfigObj)
				assert.Equal(t, tc.expectedConfig.ExpiryConfig.Duration, tc.paymentMethod.ConfigObj.ExpiryConfig.Duration)
				assert.Equal(t, tc.expectedConfig.ExpiryConfig.Unit, tc.paymentMethod.ConfigObj.ExpiryConfig.Unit)
			}
		})
	}
}

func TestPaymentMethodExpiryConfig_ToDateTime(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name           string
		config         *PaymentMethodExpiryConfig
		expectedOffset time.Duration
		unit           string
	}{
		{
			name: "SUCCESS: 30 minutes duration",
			config: &PaymentMethodExpiryConfig{
				Duration: 30,
				Unit:     constant.UnifiedPaymentExpiryUnitMinutes,
			},
			expectedOffset: 30 * time.Minute,
			unit:           "minutes",
		},
		{
			name: "SUCCESS: 1 hour duration",
			config: &PaymentMethodExpiryConfig{
				Duration: 1,
				Unit:     constant.UnifiedPaymentExpiryUnitHours,
			},
			expectedOffset: 1 * time.Hour,
			unit:           "hours",
		},
		{
			name: "SUCCESS: 24 hours duration",
			config: &PaymentMethodExpiryConfig{
				Duration: 24,
				Unit:     constant.UnifiedPaymentExpiryUnitHours,
			},
			expectedOffset: 24 * time.Hour,
			unit:           "hours",
		},
		{
			name: "SUCCESS: 7 days duration",
			config: &PaymentMethodExpiryConfig{
				Duration: 7,
				Unit:     constant.UnifiedPaymentExpiryUnitDays,
			},
			expectedOffset: 7 * 24 * time.Hour,
			unit:           "days",
		},
		{
			name: "SUCCESS: 1 day duration",
			config: &PaymentMethodExpiryConfig{
				Duration: 1,
				Unit:     constant.UnifiedPaymentExpiryUnitDays,
			},
			expectedOffset: 24 * time.Hour,
			unit:           "days",
		},
		{
			name: "SUCCESS: Default to seconds - 60 seconds",
			config: &PaymentMethodExpiryConfig{
				Duration: 60,
				Unit:     "seconds",
			},
			expectedOffset: 60 * time.Second,
			unit:           "seconds",
		},
		{
			name: "SUCCESS: Default to seconds - unknown unit",
			config: &PaymentMethodExpiryConfig{
				Duration: 120,
				Unit:     "unknown",
			},
			expectedOffset: 120 * time.Second,
			unit:           "unknown (defaults to seconds)",
		},
		{
			name: "SUCCESS: Zero duration in minutes",
			config: &PaymentMethodExpiryConfig{
				Duration: 0,
				Unit:     constant.UnifiedPaymentExpiryUnitMinutes,
			},
			expectedOffset: 0,
			unit:           "minutes",
		},
		{
			name: "SUCCESS: Zero duration in hours",
			config: &PaymentMethodExpiryConfig{
				Duration: 0,
				Unit:     constant.UnifiedPaymentExpiryUnitHours,
			},
			expectedOffset: 0,
			unit:           "hours",
		},
		{
			name: "SUCCESS: Zero duration in days",
			config: &PaymentMethodExpiryConfig{
				Duration: 0,
				Unit:     constant.UnifiedPaymentExpiryUnitDays,
			},
			expectedOffset: 0,
			unit:           "days",
		},
		{
			name: "SUCCESS: Large duration - 365 days",
			config: &PaymentMethodExpiryConfig{
				Duration: 365,
				Unit:     constant.UnifiedPaymentExpiryUnitDays,
			},
			expectedOffset: 365 * 24 * time.Hour,
			unit:           "days",
		},
		{
			name: "SUCCESS: Empty unit defaults to seconds",
			config: &PaymentMethodExpiryConfig{
				Duration: 100,
				Unit:     "",
			},
			expectedOffset: 100 * time.Second,
			unit:           "empty (defaults to seconds)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.config.ToDateTime()

			// Calculate expected time with a small tolerance for test execution time
			expectedTime := now.Add(tc.expectedOffset)

			// Allow 1 second tolerance for test execution time
			diff := result.Sub(expectedTime)
			if diff < 0 {
				diff = -diff
			}

			assert.True(t, diff <= 1*time.Second,
				"Expected time difference to be within 1 second, got %v for unit: %s", diff, tc.unit)

			// Verify the duration is approximately correct
			actualOffset := result.Sub(now)
			expectedDiff := actualOffset - tc.expectedOffset
			if expectedDiff < 0 {
				expectedDiff = -expectedDiff
			}

			assert.True(t, expectedDiff <= 1*time.Second,
				"Expected offset %v, got %v (diff: %v) for unit: %s",
				tc.expectedOffset, actualOffset, expectedDiff, tc.unit)
		})
	}
}

func TestPaymentMethodWithPivotIsCardPartnerConfigFound(t *testing.T) {
	tests := []struct {
		data       PaymentMethodWithPivot
		wantResult bool
	}{
		{
			data:       PaymentMethodWithPivot{},
			wantResult: false,
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{},
			},
			wantResult: false,
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{},
				},
			},
			wantResult: false,
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{},
					},
				},
			},
			wantResult: false,
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{{}},
						},
					},
				},
			},
			wantResult: true,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantResult, test.data.IsCardPartnerConfigFound())
	}
}

func TestPaymentMethodWithPivotGetCardPartnerConfigForOnlineTravelAgent(t *testing.T) {
	def := config.VCCTerminalDefaultConfig{
		CardTypes:          []string{"CREDIT"}, // NOSONAR
		AcquirerMerchantID: "TESTING12345",     // NOSONAR
		PrincipalAvailable: []string{"VISA"},   // NOSONAR
	}

	tests := []struct {
		data       PaymentMethodWithPivot
		wantResult map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj
	}{
		{
			data: PaymentMethodWithPivot{},
			wantResult: map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
				c.DefaultConfig: {
					CardTypes:          def.CardTypes,
					AcquirerMerchantID: def.AcquirerMerchantID,
					PrincipalAvailable: def.PrincipalAvailable,
					SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
						AllowedBinNumbers: []string{"ALL"}, // NOSONAR
					},
				},
			},
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					VirtualTerminalConfig: &paymentMethodModel.VirtualTerminalConfig{
						AcquirerMerchantID: "TEST999988",           // NOSONAR
						AllowedBinNumbers:  []string{"123456"},     // NOSONAR
						CardTypes:          []string{"DEBIT"},      // NOSONAR
						PrincipalAvailable: []string{"MASTERCARD"}, // NOSONAR
					},
				},
			},
			wantResult: map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
				c.DefaultConfig: {
					CardTypes:          []string{"DEBIT"},      // NOSONAR
					AcquirerMerchantID: "TEST999988",           // NOSONAR
					PrincipalAvailable: []string{"MASTERCARD"}, // NOSONAR
					SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
						AllowedBinNumbers: []string{"123456"}, // NOSONAR
					},
				},
			},
		},
		{
			data: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					VirtualTerminalConfig: &paymentMethodModel.VirtualTerminalConfig{
						AcquirerMerchantID: "TEST999988",           // NOSONAR
						AllowedBinNumbers:  []string{"123456"},     // NOSONAR
						CardTypes:          []string{"DEBIT"},      // NOSONAR
						PrincipalAvailable: []string{"MASTERCARD"}, // NOSONAR
					},
					PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
						Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
							Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
								{
									TravelAgentCode:    "PIVOT",                     // NOSONAR
									PartnerProcessor:   "MPGS",                      // NOSONAR
									AcquirerMerchantID: "PIVOT_00001",               // NOSONAR
									IsActive:           true,                        // NOSONAR
									CardTypes:          []string{"DEBIT", "CREDIT"}, // NOSONAR
									PrincipalAvailable: []string{"JCB"},             // NOSONAR
									SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
										AllowedBinNumbers: []string{"987654"}, // NOSONAR
									},
								},
							},
						},
					},
				},
			},
			wantResult: map[string]*paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
				c.DefaultConfig: {
					CardTypes:          []string{"DEBIT"},      // NOSONAR
					AcquirerMerchantID: "TEST999988",           // NOSONAR
					PrincipalAvailable: []string{"MASTERCARD"}, // NOSONAR
					SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
						AllowedBinNumbers: []string{"123456"}, // NOSONAR
					},
				},
				"PIVOT": {
					TravelAgentCode:    "PIVOT",                     // NOSONAR
					PartnerProcessor:   "MPGS",                      // NOSONAR
					AcquirerMerchantID: "PIVOT_00001",               // NOSONAR
					IsActive:           true,                        // NOSONAR
					CardTypes:          []string{"DEBIT", "CREDIT"}, // NOSONAR
					PrincipalAvailable: []string{"JCB"},             // NOSONAR
					SupportedUseCase: &paymentMethodModel.CardSupportedUseCase{
						AllowedBinNumbers: []string{"987654"}, // NOSONAR
					},
				},
			},
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantResult, test.data.GetCardPartnerConfigForOnlineTravelAgent(def))
	}
}

func TestPaymentMethodWithPivotEnableSplitCardPayment(t *testing.T) {
	tests := []struct {
		name       string
		input      PaymentMethodWithPivot
		wantResult bool
	}{
		{
			name:       "should return false when MerchantConfigObj is nil",
			input:      PaymentMethodWithPivot{},
			wantResult: false,
		},
		{
			name: "should return false when SplitCardPaymentConfig is nil",
			input: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{},
			},
			wantResult: false,
		},
		{
			name: "should return false when Enabled is false",
			input: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					SplitCardPaymentConfig: &paymentMethodModel.SplitCardPaymentConfig{
						Enabled: false,
					},
				},
			},
			wantResult: false,
		},
		{
			name: "should return true when Enabled is true",
			input: PaymentMethodWithPivot{
				MerchantConfigObj: &PaymentMethodMerchantConfigObject{
					SplitCardPaymentConfig: &paymentMethodModel.SplitCardPaymentConfig{
						Enabled:         true,
						ActiveProcessor: "MPGS",
						Processors: map[string]paymentMethodModel.CardPartnerProcessorConfig{
							"MPGS": {Limit: 2000000000},
						},
					},
				},
			},
			wantResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.input.EnableSplitCardPayment())
		})
	}
}

func TestPaymentMethodMerchantConfigObject_GetCardAcquirer(t *testing.T) {
	tests := []struct {
		name       string
		config     *PaymentMethodMerchantConfigObject
		mid        string
		wantResult string
	}{
		{
			name:       "PartnerConfig is nil",
			config:     &PaymentMethodMerchantConfigObject{},
			mid:        "MID001",
			wantResult: "",
		},
		{
			name: "Card is nil",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{},
			},
			mid:        "MID001",
			wantResult: "",
		},
		{
			name: "Card items is empty",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{},
					},
				},
			},
			mid:        "MID001",
			wantResult: "",
		},
		{
			name: "MID not found in items",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{AcquirerMerchantID: "MID002", Acquirer: "BCA"},
						},
					},
				},
			},
			mid:        "MID001",
			wantResult: "",
		},
		{
			name: "MID found - returns acquirer",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{AcquirerMerchantID: "MID001", Acquirer: "BCA"},
						},
					},
				},
			},
			mid:        "MID001",
			wantResult: "BCA",
		},
		{
			name: "MID found in multiple items - returns matching acquirer",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{AcquirerMerchantID: "MID001", Acquirer: "BCA"},
							{AcquirerMerchantID: "MID002", Acquirer: "MANDIRI"},
							{AcquirerMerchantID: "MID003", Acquirer: "BNI"},
						},
					},
				},
			},
			mid:        "MID002",
			wantResult: "MANDIRI",
		},
		{
			name: "Empty MID input",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{AcquirerMerchantID: "", Acquirer: "BCA"},
						},
					},
				},
			},
			mid:        "",
			wantResult: "BCA",
		},
		{
			name: "Empty acquirer value",
			config: &PaymentMethodMerchantConfigObject{
				PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
					Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
						Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
							{AcquirerMerchantID: "MID001", Acquirer: ""},
						},
					},
				},
			},
			mid:        "MID001",
			wantResult: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetCardAcquirer(tt.mid)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}
