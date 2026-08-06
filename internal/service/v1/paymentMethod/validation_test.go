package paymentMethodService

import (
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/stretchr/testify/assert"
)

func TestValidateCardPartnerConfig(t *testing.T) {

	travelAgents := config.MStrStr{"TEST": "Test.com"} // NOSONAR

	errMustEnable3dsBypass := errors.New("acquirer merchant ID (CIT/MIT) must enable 3DS bypass")
	errMustNotEnable3dsBypass := errors.New("acquirer merchant ID (CIT) must not enable 3DS bypass")

	tests := []struct {
		name              string
		paymentMethodType string
		config            model.SetupPaymentMethodPartnerConfigForCardObj
		dupChecker        *uniqueData
		wantError         error
	}{
		{
			name:      "SUCCESS: General config", // NOSONAR
			config:    model.SetupPaymentMethodPartnerConfigForCardObj{},
			wantError: nil,
		},
		{
			name: "SUCCESS: OTA config", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				TravelAgentCode: "TEST", // NOSONAR
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:    true,
					AllowForeignCard:  true,
					AllowedBinNumbers: []string{"444000", "55500000", "55500000111"},
				},
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Card-funded payout", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeCIT,
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Recurring payment", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				RecurringType: constant.CardTransactionTypeMIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: true,
				},
			},
			wantError: nil,
		},
		{
			name: "SUCCESS: Split payment with 3ds bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				SplitPaymentType: constant.CardTransactionTypeCIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:       true,
					AllowExternalThreeDs: true,
				},
			},
			wantError: nil,
		},
		{
			name: "ERROR: OTA config must include supported use cases", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				TravelAgentCode: "TEST", // NOSONAR
			},
			wantError: errors.New("partner configuration for OTA must allow non-3ds transactions, foreign cards and include the allowed bin numbers"),
		},
		{
			name: "ERROR: Invalid BIN number length", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowedBinNumbers: []string{"4440"},
				},
			},
			wantError: errors.New("invalid BIN number length"),
		},
		{
			name: "ERROR: Travel agent code not found",
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				TravelAgentCode: "ABCD", // NOSONAR
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:    true,
					AllowForeignCard:  true,
					AllowedBinNumbers: []string{"444000"},
				},
			},
			wantError: errors.New("travel agent code ABCD not found"),
		},
		{
			name: "ERROR: Duplicate config",
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				TravelAgentCode: "TEST", // NOSONAR
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:    true,
					AllowForeignCard:  true,
					AllowedBinNumbers: []string{"444000"},
				},
			},
			dupChecker: &uniqueData{travelAgents: map[string]bool{"TEST": true}},
			wantError:  errors.New("duplicate configuration found for travel agent code TEST"),
		},
		{
			name: "ERROR: Card funded payout CIT type must not enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeCIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: true,
				},
			},
			wantError: errMustNotEnable3dsBypass,
		},
		{
			name: "ERROR: Card funded payout MIT type must enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeMIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: false,
				},
			},
			wantError: errMustEnable3dsBypass,
		},
		{
			name: "ERROR: Recurring payment CIT type must not enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				RecurringType: constant.CardTransactionTypeCIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: true,
				},
			},
			wantError: errMustNotEnable3dsBypass,
		},
		{
			name: "ERROR: Recurring payment MIT type must enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				RecurringType: constant.CardTransactionTypeMIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: false,
				},
			},
			wantError: errMustEnable3dsBypass,
		},
		{
			name: "ERROR: Split payment CIT type must enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				SplitPaymentType: constant.CardTransactionTypeCIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: false,
				},
			},
			wantError: errMustEnable3dsBypass,
		},
		{
			name: "ERROR: Split payment MIT type must enable 3DS bypass", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				SplitPaymentType: constant.CardTransactionTypeMIT,
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds: false,
				},
			},
			wantError: errMustEnable3dsBypass,
		},
		{
			name: "ERROR: Split payment Should enable external 3DS", // NOSONAR
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				SplitPaymentType:   constant.CardTransactionTypeCIT,
				AcquirerMerchantID: "ABC", // NOSONAR
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:       true,
					AllowExternalThreeDs: false,
				},
			},
			wantError: errors.New("acquirer merchant ID ABC is required to enable external 3DS for specific use cases"),
		},
		{
			name:              "ERROR: Empty travel agent code on virtual terminal type",
			paymentMethodType: paymentConst.PAYMENT_METHOD_VIRTUAL_TERMINAL,
			config: model.SetupPaymentMethodPartnerConfigForCardObj{
				TravelAgentCode: "",
			},
			wantError: errors.New("travel agent code is required for the virtual terminal payment method"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.dupChecker == nil {
				test.dupChecker = &uniqueData{travelAgents: map[string]bool{}}
			}
			assert.Equal(t, test.wantError, validateCardPartnerConfig(test.paymentMethodType, test.config, test.dupChecker, travelAgents))
		})
	}
}

func TestValidateSplitCardPaymentConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  model.SetupPaymentMethodPartnerConfigForCardRequest
		wantErr error
	}{
		{
			name:    "SUCCESS: empty items and nil split config",
			config:  model.SetupPaymentMethodPartnerConfigForCardRequest{},
			wantErr: nil,
		},
		{
			name: "SUCCESS: neither split config nor split items present",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{PartnerProcessor: "MPGS", AcquirerMerchantID: "MID001"},
				},
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: split config set with matching CIT and MIT items for active processor",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors: map[string]model.CardPartnerProcessorConfig{
						"MPGS": {Limit: 10000},
					},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "MPGS", IsActive: true},
					{SplitPaymentType: constant.CardTransactionTypeMIT, PartnerProcessor: "MPGS", IsActive: true},
				},
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS: split config with multiple processors but active processor has matching items",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors: map[string]model.CardPartnerProcessorConfig{
						"MPGS": {Limit: 10000},
						"CYBS": {Limit: 5000},
					},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "MPGS", IsActive: true},
					{SplitPaymentType: constant.CardTransactionTypeMIT, PartnerProcessor: "MPGS", IsActive: true},
				},
			},
			wantErr: nil,
		},
		{
			name: "ERROR: split config not set but CIT and MIT split types defined in partner config",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "MPGS"},
				},
			},
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("invalid config: split card payment is not set, but CIT or MIT split types are defined in partner config")),
		},
		{
			name: "ERROR: split config set but CIT and MIT split types missing in partner config",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors: map[string]model.CardPartnerProcessorConfig{
						"MPGS": {Limit: 10000},
					},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{PartnerProcessor: "MPGS", AcquirerMerchantID: "MID001"},
				},
			},
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("invalid config: split card payment is set, but CIT or MIT split types are missing in the partner config")),
		},
		{
			name: "ERROR: missing processor limit for active processor",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors:      map[string]model.CardPartnerProcessorConfig{},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "MPGS", IsActive: true},
					{SplitPaymentType: constant.CardTransactionTypeMIT, PartnerProcessor: "MPGS", IsActive: true},
				},
			},
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("missing processor limit for active processor")),
		},
		{
			name: "ERROR: missing partner config for active processor",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors: map[string]model.CardPartnerProcessorConfig{
						"MPGS": {Limit: 10000},
					},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "CYBS", IsActive: true},
					{SplitPaymentType: constant.CardTransactionTypeMIT, PartnerProcessor: "CYBS", IsActive: true},
				},
			},
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("missing partner config for active processor")),
		},
		{
			name: "ERROR: split payment item exists but is not active for active processor",
			config: model.SetupPaymentMethodPartnerConfigForCardRequest{
				SplitCardPaymentConfig: &model.SplitCardPaymentConfig{
					ActiveProcessor: "MPGS",
					Processors: map[string]model.CardPartnerProcessorConfig{
						"MPGS": {Limit: 10000},
					},
				},
				Items: []model.SetupPaymentMethodPartnerConfigForCardObj{
					{SplitPaymentType: constant.CardTransactionTypeCIT, PartnerProcessor: "MPGS", IsActive: false},
					{SplitPaymentType: constant.CardTransactionTypeMIT, PartnerProcessor: "MPGS", IsActive: false},
				},
			},
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("missing partner config for active processor")),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, validateSplitCardPaymentConfig(&tt.config))
		})
	}
}

func TestShouldDisableBypass3DS(t *testing.T) {
	tests := []struct {
		name    string
		cardCfg model.SetupPaymentMethodPartnerConfigForCardObj
		want    bool
	}{
		{
			name:    "no CIT type - CardFundedPayoutType is MIT",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{CardFundedPayoutType: constant.CardTransactionTypeMIT},
			want:    false,
		},
		{
			name:    "no CIT type - all empty",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{},
			want:    false,
		},
		{
			name: "CIT via CardFundedPayoutType - SupportedUseCase allows bypass",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeCIT,
				SupportedUseCase:     &model.CardSupportedUseCase{AllowBypass3Ds: true},
			},
			want: true,
		},
		{
			name: "CIT via CardFundedPayoutType - SupportedUseCase disallows bypass",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeCIT,
				SupportedUseCase:     &model.CardSupportedUseCase{AllowBypass3Ds: false},
			},
			want: false,
		},
		{
			name: "CIT via CardFundedPayoutType - nil SupportedUseCase",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{
				CardFundedPayoutType: constant.CardTransactionTypeCIT,
			},
			want: false,
		},
		{
			name: "CIT via RecurringType - allows bypass",
			cardCfg: model.SetupPaymentMethodPartnerConfigForCardObj{
				RecurringType:    constant.CardTransactionTypeCIT,
				SupportedUseCase: &model.CardSupportedUseCase{AllowBypass3Ds: true},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldDisableBypass3DS(tt.cardCfg))
		})
	}
}

func TestValidateCardTransactionMIDPairConfig(t *testing.T) {
	tests := []struct {
		name    string
		label   string
		configs []cardConfig
		keyFn   func(cardConfig) string
		wantErr error
	}{
		{
			name:  "ERROR:Recurring Payment-MID MIT not found",
			label: "recurring payment",
			configs: []cardConfig{
				{RecurringType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
			},
			keyFn:   func(c cardConfig) string { return c.PartnerProcessor + "-" + c.RecurringType },
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("invalid MID configuration for recurring payment. MPGS-MIT configuration not found")),
		},
		{
			name:  "ERROR:Card-funded Payout-Duplicate config",
			label: "card-funded payout",
			configs: []cardConfig{
				{CardFundedPayoutType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
				{CardFundedPayoutType: constant.CardTransactionTypeMIT, AcquirerMerchantID: "DEF", PartnerProcessor: "MPGS", IsActive: true},
				{CardFundedPayoutType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
			},
			keyFn:   func(c cardConfig) string { return c.PartnerProcessor + "-" + c.CardFundedPayoutType },
			wantErr: pkgErr.New(response.HttpErrRequest, errors.New("invalid MID configuration for card-funded payout: duplicate configuration for MPGS-CIT type")),
		},
		{
			name: "SUCCESS:Recurring Payment",
			configs: []cardConfig{
				{RecurringType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
				{RecurringType: constant.CardTransactionTypeMIT, AcquirerMerchantID: "DEF", PartnerProcessor: "MPGS", IsActive: true},
			},
			keyFn:   func(c cardConfig) string { return c.PartnerProcessor + "-" + c.RecurringType },
			wantErr: nil,
		},
		{
			name: "SUCCESS:Card-funded Payout",
			configs: []cardConfig{
				{CardFundedPayoutType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
				{CardFundedPayoutType: constant.CardTransactionTypeMIT, AcquirerMerchantID: "DEF", PartnerProcessor: "MPGS", IsActive: true},
			},
			keyFn:   func(c cardConfig) string { return c.PartnerProcessor + "-" + c.CardFundedPayoutType },
			wantErr: nil,
		},
		{
			name: "SUCCESS:Split Payment",
			configs: []cardConfig{
				{SplitPaymentType: constant.CardTransactionTypeCIT, AcquirerMerchantID: "ABC", PartnerProcessor: "MPGS", IsActive: true},
				{SplitPaymentType: constant.CardTransactionTypeMIT, AcquirerMerchantID: "DEF", PartnerProcessor: "MPGS", IsActive: true},
			},
			keyFn:   func(c cardConfig) string { return c.PartnerProcessor + "-" + c.SplitPaymentType },
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantErr, validateCardTransactionMIDPairConfig(tt.label, tt.configs, tt.keyFn))
		})
	}
}

func TestSetupValueForSpecificPaymentType(t *testing.T) {
	tests := []struct {
		name              string
		paymentMethodType string
		config            *model.SetupPaymentMethodPartnerConfigForCardObj
		wantResult        model.SetupPaymentMethodPartnerConfigForCardObj
	}{
		{
			name:       "Nil item config",
			config:     &model.SetupPaymentMethodPartnerConfigForCardObj{},
			wantResult: model.SetupPaymentMethodPartnerConfigForCardObj{},
		},
		{
			name:              "Virtual terminal",
			paymentMethodType: paymentConst.PAYMENT_METHOD_VIRTUAL_TERMINAL,
			config:            &model.SetupPaymentMethodPartnerConfigForCardObj{},
			wantResult: model.SetupPaymentMethodPartnerConfigForCardObj{
				SupportedUseCase: &model.CardSupportedUseCase{
					AllowBypass3Ds:   true,
					AllowForeignCard: true,
				},
			},
		},
		{
			name:              "Credit Card",
			paymentMethodType: paymentConst.PAYMENT_METHOD_CREDIT_CARD,
			config:            &model.SetupPaymentMethodPartnerConfigForCardObj{},
			wantResult:        model.SetupPaymentMethodPartnerConfigForCardObj{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, setupValueForSpecificPaymentType(tt.paymentMethodType, tt.config))
		})
	}
}
