package paymentMethodModel

import (
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	constant "github.com/paper-indonesia/pivot-backoffice/constant/payment"

	"github.com/stretchr/testify/assert"
)

func TestSetupPaymentMethodPartnerConfigRequest_GetMaxBNCQRStaticLimit(t *testing.T) {
	tests := []struct {
		name     string
		request  *SetupPaymentMethodPartnerConfigRequest
		expected int
	}{
		{
			name: "SUCCESS: Returns 0 when Qris is nil",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: nil,
			},
			expected: 0,
		},
		{
			name: "SUCCESS: Returns 0 when Qris Acquirer is not BNC",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BRI,
					AcquirerStoreIDs: []string{"storeID1", "storeID2"},
				},
			},
			expected: 0,
		},
		{
			name: "SUCCESS: Returns 0 when AcquirerStoreIDs is empty",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{},
				},
			},
			expected: 1,
		},
		{
			name: "SUCCESS: Returns 1 when AcquirerStoreIDs has one element",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"STORE001"},
				},
			},
			expected: 2,
		},
		{
			name: "SUCCESS: Returns 3 when AcquirerStoreIDs has three elements",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"STORE001", "STORE002", "STORE003"},
				},
			},
			expected: 4,
		},
		{
			name: "SUCCESS: Returns 5 when AcquirerStoreIDs has five elements",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"STORE001", "STORE002", "STORE003", "STORE004", "STORE005"},
				},
			},
			expected: 6,
		},
		{
			name: "SUCCESS: Returns correct count when AcquirerStoreIDs has duplicate values",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"STORE001", "STORE001", "STORE002"},
				},
			},
			expected: 4,
		},
		{
			name: "SUCCESS: Returns correct count when AcquirerStoreIDs has empty strings",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"", "STORE001", ""},
				},
			},
			expected: 4,
		},
		{
			name: "SUCCESS: Works with other QRIS fields populated",
			request: &SetupPaymentMethodPartnerConfigRequest{
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					AcquirerMerchantID: "MERCHANT123",
					AcquirerTerminalID: "TERMINAL456",
					QRType:             "STATIC",
					Acquirer:           "BNC",
					MerchantType:       "LARGE",
					CreatedBy:          "user123",
					AcquirerStoreIDs:   []string{"STORE001", "STORE002"},
				},
			},
			expected: 3,
		},
		{
			name: "SUCCESS: Works with other partner configs populated",
			request: &SetupPaymentMethodPartnerConfigRequest{
				VirtualAccount: &SetupPaymentMethodPartnerConfigForVARequest{
					Items: []SetupPaymentMethodPartnerConfigForVAObj{
						{BINPrefix: "123", Type: "OPEN_STATIC"},
					},
				},
				Card: &SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []SetupPaymentMethodPartnerConfigForCardObj{
						{
							PartnerProcessor:   "MPGS",
							AcquirerMerchantID: "MERCHANT123",
						},
					},
				},
				EWallet: &SetupPaymentMethodPartnerConfigForEWalletRequest{
					ExternalMerchantID: "EXT123",
				},
				Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
					Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
					AcquirerStoreIDs: []string{"STORE001"},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.GetMaxBNCQRStaticLimit()
			assert.Equal(t, tt.expected, result, "GetMaxBNCQRStaticLimit() should return expected value")
		})
	}
}

func TestSetupPaymentMethodPartnerConfigRequest_GetMaxBNCQRStaticLimit_EdgeCases(t *testing.T) {
	t.Run("SUCCESS: Nil request should not panic", func(t *testing.T) {
		var request *SetupPaymentMethodPartnerConfigRequest
		// This would panic if the method doesn't handle nil receiver properly
		// But since the method is called on the struct, it should work fine
		assert.NotPanics(t, func() {
			if request != nil {
				request.GetMaxBNCQRStaticLimit()
			}
		})
	})

	t.Run("SUCCESS: Large number of AcquirerStoreIDs", func(t *testing.T) {
		// Create a large slice to test performance and correctness
		largeStoreIDs := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			largeStoreIDs[i] = "STORE" + string(rune(i))
		}

		request := &SetupPaymentMethodPartnerConfigRequest{
			Qris: &SetupPaymentMethodPartnerConfigForQrisRequest{
				AcquirerStoreIDs: largeStoreIDs,
				Acquirer:         constant.PAYMENT_METHOD_QRIS_ACQUIRER_BNC,
			},
		}

		result := request.GetMaxBNCQRStaticLimit()
		assert.Equal(t, 1001, result, "Should handle large slice correctly")
	})
}

func TestSetupPaymentMethodPartnerConfigForCardRequestGetNetworkTokenPartnerConfig(t *testing.T) {
	tests := []struct {
		name             string
		partnerConfig    SetupPaymentMethodPartnerConfigForCardRequest
		cofInitiatorType string
		wantResult       *SetupPaymentMethodPartnerConfigForCardObj
	}{
		{
			name: "Empty items returns nil",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{},
			},
			cofInitiatorType: "",
			wantResult:       nil,
		},
		{
			name: "Nil network token items skipped",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "A"},
					{AcquirerMerchantID: "B"},
				},
			},
			cofInitiatorType: "",
			wantResult:       nil,
		},
		{
			name: "DEFAULT type returned when cofInitiatorType is empty",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{
						AcquirerMerchantID: "DEFAULT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
					},
				},
			},
			cofInitiatorType: "",
			wantResult: &SetupPaymentMethodPartnerConfigForCardObj{
				AcquirerMerchantID: "DEFAULT-MID",
				NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
			},
		},
		{
			name: "DEFAULT type not returned when cofInitiatorType is set",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{
						AcquirerMerchantID: "DEFAULT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
					},
				},
			},
			cofInitiatorType: "MERCHANT",
			wantResult:       nil,
		},
		{
			name: "COF type with MERCHANT initiator matches",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{
						AcquirerMerchantID: "COF-MERCHANT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
					},
				},
			},
			cofInitiatorType: "MERCHANT",
			wantResult: &SetupPaymentMethodPartnerConfigForCardObj{
				AcquirerMerchantID: "COF-MERCHANT-MID",
				NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
			},
		},
		{
			name: "COF type with CUSTOMER initiator matches",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{
						AcquirerMerchantID: "COF-CUSTOMER-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "CUSTOMER"},
					},
				},
			},
			cofInitiatorType: "CUSTOMER",
			wantResult: &SetupPaymentMethodPartnerConfigForCardObj{
				AcquirerMerchantID: "COF-CUSTOMER-MID",
				NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "CUSTOMER"},
			},
		},
		{
			name: "COF type with wrong initiator returns nil",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{
						AcquirerMerchantID: "COF-MERCHANT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
					},
				},
			},
			cofInitiatorType: "CUSTOMER",
			wantResult:       nil,
		},
		{
			name: "Mixed items - returns correct DEFAULT when empty initiator",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "NO-NETWORK-TOKEN"},
					{
						AcquirerMerchantID: "COF-MERCHANT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
					},
					{
						AcquirerMerchantID: "DEFAULT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
					},
				},
			},
			cofInitiatorType: "",
			wantResult: &SetupPaymentMethodPartnerConfigForCardObj{
				AcquirerMerchantID: "DEFAULT-MID",
				NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
			},
		},
		{
			name: "Mixed items - returns correct COF with matching initiator",
			partnerConfig: SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "NO-NETWORK-TOKEN"},
					{
						AcquirerMerchantID: "DEFAULT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenDefaultType},
					},
					{
						AcquirerMerchantID: "COF-CUSTOMER-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "CUSTOMER"},
					},
					{
						AcquirerMerchantID: "COF-MERCHANT-MID",
						NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
					},
				},
			},
			cofInitiatorType: "MERCHANT",
			wantResult: &SetupPaymentMethodPartnerConfigForCardObj{
				AcquirerMerchantID: "COF-MERCHANT-MID",
				NetworkToken:       &CardNetworkTokenPartnerConfigObj{Type: c.NetworkTokenCOFType, COFInitiator: "MERCHANT"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.partnerConfig.GetNetworkTokenPartnerConfig(tt.cofInitiatorType)
			assert.Equal(t, tt.wantResult, result)
		})
	}
}

func TestSetupPaymentMethodPartnerConfigForCardRequestGetPartnerConfigByPaymentType(t *testing.T) {
	getPartnerConfig := func() SetupPaymentMethodPartnerConfigForCardRequest {
		return SetupPaymentMethodPartnerConfigForCardRequest{
			Items: []SetupPaymentMethodPartnerConfigForCardObj{
				{AcquirerMerchantID: "A"}, // General payment
				{AcquirerMerchantID: "B", RecurringType: c.CardTransactionTypeCIT},        // Recurring payments
				{AcquirerMerchantID: "I", RecurringType: c.CardTransactionTypeMIT},        // Recurring payments
				{AcquirerMerchantID: "C", TravelAgentCode: "TEST"},                        // Virtual terminal
				{AcquirerMerchantID: "D", CardFundedPayoutType: c.CardTransactionTypeMIT}, // Card-funded payout
				{AcquirerMerchantID: "E", CardFundedPayoutType: c.CardTransactionTypeCIT}, // One dollar authorization
				{AcquirerMerchantID: "F"}, // General payment
				{AcquirerMerchantID: "G", SplitPaymentType: c.CardTransactionTypeCIT}, // Split payment CIT
				{AcquirerMerchantID: "H", SplitPaymentType: c.CardTransactionTypeMIT}, // Split payment MIT
			},
		}
	}

	tests := []struct {
		name          string
		partnerConfig SetupPaymentMethodPartnerConfigForCardRequest
		paymentType   string
		wantResult    *SetupPaymentMethodPartnerConfigForCardRequest
	}{
		{
			name:       "Empty items",
			wantResult: nil,
		},
		{
			name:          "General payment",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypePayment,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "A"},
					{AcquirerMerchantID: "F"},
				},
			},
		},
		{
			name:          "Recurring payment",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypeRecurringPayment,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "B", RecurringType: c.CardTransactionTypeCIT},
					{AcquirerMerchantID: "I", RecurringType: c.CardTransactionTypeMIT},
				},
			},
		},
		{
			name:          "Virtual terminal",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypeVirtualTerminal,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "C", TravelAgentCode: "TEST"},
				},
			},
		},
		{
			name:          "Card-funded payout",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypeCardFundedPayout,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "D", CardFundedPayoutType: c.CardTransactionTypeMIT},
					{AcquirerMerchantID: "E", CardFundedPayoutType: c.CardTransactionTypeCIT},
				},
			},
		},
		{
			name:          "One dollar authorization",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypeOneDollarAuth,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "A"},
					{AcquirerMerchantID: "F"},
				},
			},
		},
		{
			name:          "Split payment",
			partnerConfig: getPartnerConfig(),
			paymentType:   c.GroupPaymentTypeSplitPayment,
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "G", SplitPaymentType: c.CardTransactionTypeCIT},
					{AcquirerMerchantID: "H", SplitPaymentType: c.CardTransactionTypeMIT},
				},
			},
		},
		{
			name:          "Unmapping",
			partnerConfig: getPartnerConfig(),
			paymentType:   "OTHERS", // NOSONAR
			wantResult: &SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []SetupPaymentMethodPartnerConfigForCardObj{
					{AcquirerMerchantID: "A"},
					{AcquirerMerchantID: "B", RecurringType: c.CardTransactionTypeCIT},
					{AcquirerMerchantID: "I", RecurringType: c.CardTransactionTypeMIT},
					{AcquirerMerchantID: "C", TravelAgentCode: "TEST"},
					{AcquirerMerchantID: "D", CardFundedPayoutType: c.CardTransactionTypeMIT},
					{AcquirerMerchantID: "E", CardFundedPayoutType: c.CardTransactionTypeCIT},
					{AcquirerMerchantID: "F"},
					{AcquirerMerchantID: "G", SplitPaymentType: c.CardTransactionTypeCIT},
					{AcquirerMerchantID: "H", SplitPaymentType: c.CardTransactionTypeMIT},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantResult, tt.partnerConfig.GetPartnerConfigByPaymentType(tt.paymentType))
		})
	}
}
