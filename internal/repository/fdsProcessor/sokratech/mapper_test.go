package sokratech

import (
	"testing"
	"time"

	common "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	fdscommon "github.com/paper-indonesia/pivot-backoffice/internal/model/fdsProcessor/fdsCommon"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/sokratech"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/shopspring/decimal"

	"github.com/stretchr/testify/assert"
)

func TestToPayoutWorkflowRequest(t *testing.T) {
	data := fdscommon.AssessPayoutTransactionRequest{
		Merchant: fdscommon.Merchant{
			ID:        "3315935d-5e6a-40ca-b316-7f85fa8b8a30", // NOSONAR
			Name:      "Test Merchant",                        // NOSONAR
			RiskLevel: "LOW",                                  // NOSONAR
		},
		Transaction: fdscommon.Transaction{
			ID:                "e3cdc973-31c4-409c-bcd2-f7a5efb0720b", // NOSONAR
			ClientReferenceID: "REF00011110100293",                    // NOSONAR
			Amount: common.Amount2{
				Value:    12_250, // NOSONAR
				Currency: "IDR",  // NOSONAR
			},
			CreatedAt:   time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC), // NOSONAR
			CreatedFrom: "OPEN_API",                                   // NOSONAR
		},
		Destination: fdscommon.PayoutDestination{
			BankCode:      "002",          // NOSONAR
			AccountNumber: "029930495555", // NOSONAR
			AccountName:   "John Doe",     // NOSONAR
		},
	}
	wantPayload := model.PayoutWorkflowRequest{
		Merchant: model.Merchant{
			ID:        "3315935d-5e6a-40ca-b316-7f85fa8b8a30", // NOSONAR
			Name:      "Test Merchant",                        // NOSONAR
			RiskLevel: "LOW",                                  // NOSONAR
		},
		Transaction: model.Transaction{
			ID:                "e3cdc973-31c4-409c-bcd2-f7a5efb0720b", // NOSONAR
			ClientReferenceID: "REF00011110100293",                    // NOSONAR
			Amount: common.Amount2{
				Value:    12_250, // NOSONAR
				Currency: "IDR",  // NOSONAR
			},
			CreatedAt:   time.Date(2026, 2, 3, 2, 34, 2, 0, time.UTC), // NOSONAR
			CreatedFrom: "OPEN_API",                                   // NOSONAR
		},
		Destination: model.PayoutDestination{
			BankCode:                "002",          // NOSONAR
			AccountNumber:           "029930495555", // NOSONAR
			AccountName:             "John Doe",     // NOSONAR
			AccountNumberTypeNumber: 29930495555,    // NOSONAR
		},
	}
	assert.Equal(t, wantPayload, toPayoutWorkflowRequest(data))
}

func TestToPaymentWorkflowRequest(t *testing.T) {

	t.Run("card payment with all fields populated", func(t *testing.T) {
		createdAt := time.Date(2026, 2, 3, 10, 30, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 2, 3, 10, 35, 0, 0, time.UTC)

		data := &fdscommon.CheckRequest{
			Partner: fdscommon.PartnerCheck{
				ID:        "merchant-123",               // NOSONAR
				Company:   util.ValueToPtr("Test Corp"), // NOSONAR
				RiskLevel: "MEDIUM",                     // NOSONAR
			},
			Customer: fdscommon.CustomerCheck{
				FirstName: util.ValueToPtr("John"),             // NOSONAR
				Email:     util.ValueToPtr("john@example.com"), // NOSONAR
				Phone:     util.ValueToPtr("+6281234567890"),   // NOSONAR
			},
			Transaction: fdscommon.TransactionCheck{
				ID:                "order-456",                                      // NOSONAR
				ClientReferenceID: "client-ref-789",                                 // NOSONAR
				OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(100000.50)), // NOSONAR
				OrderCurrency:     util.ValueToPtr("IDR"),                           // NOSONAR
				CreatedAt:         createdAt,                                        // NOSONAR
				UpdatedAt:         updatedAt,                                        // NOSONAR
			},
			Payment: fdscommon.PaymentCheck{
				MethodType:       "CREDIT_CARD",              // NOSONAR
				Fingerprint:      "fp-123",                   // NOSONAR
				MaskedCardNumber: "411111******1111",         // NOSONAR
				CardBrand:        "VISA",                     // NOSONAR
				CardCountryCode:  "ID",                       // NOSONAR
				CardType:         "CREDIT",                   // NOSONAR
				CardIssuing:      "Bank ABC",                 // NOSONAR
				ThreeDsEci:       util.ValueToPtr("05"),      // NOSONAR
				AuthCode:         util.ValueToPtr("AUTH123"), // NOSONAR
				CvvResultCode:    util.ValueToPtr("M"),       // NOSONAR
			},
			Device: fdscommon.DeviceCheck{
				IPType:    "v4",                             // NOSONAR
				IPAddress: util.ValueToPtr("192.168.1.100"), // NOSONAR
			},
			Custom: &fdscommon.CustomCheck{
				Number:        util.ValueToPtr("MID-999"),        // NOSONAR
				AcquiringName: util.ValueToPtr("Acquiring Bank"), // NOSONAR
			},
		}

		expected := model.PaymentWorkflowRequest{
			Merchant: model.Merchant{
				ID:        "merchant-123", // NOSONAR
				Name:      "Test Corp",    // NOSONAR
				RiskLevel: "MEDIUM",       // NOSONAR
			},
			Customer: model.Customer{
				Name:        "John",             // NOSONAR
				Email:       "john@example.com", // NOSONAR
				PhoneNumber: "+6281234567890",   // NOSONAR
			},
			Transaction: model.Transaction{
				ID:                "order-456",      // NOSONAR
				ClientReferenceID: "client-ref-789", // NOSONAR
				Amount: common.Amount2{
					Value:    100000.50, // NOSONAR
					Currency: "IDR",     // NOSONAR
				},
				CreatedAt: createdAt, // NOSONAR
				UpdatedAt: updatedAt, // NOSONAR
			},
			PaymentMethod: model.PaymentMethod{
				Type: "CARD", // NOSONAR
				Card: &model.PaymentMethodTypeCard{
					CardFingerprint: "fp-123",           // NOSONAR
					CardNumber:      "411111******1111", // NOSONAR
					CardBrand:       "VISA",             // NOSONAR
					CardCountryCode: "ID",               // NOSONAR
					CardType:        "CREDIT",           // NOSONAR
					IssuerName:      "Bank ABC",         // NOSONAR
					ECICode:         "05",               // NOSONAR
					ApprovalCode:    "AUTH123",          // NOSONAR
					CvvCode:         "M",                // NOSONAR
					BankMerchantID:  "MID-999",          // NOSONAR
					AcquirerName:    "Acquiring Bank",   // NOSONAR
				},
			},
			Device: model.Device{
				IPType:    "v4",            // NOSONAR
				IPAddress: "192.168.1.100", // NOSONAR
			},
			Metadata: map[string]any{},
		}

		result := toPaymentWorkflowRequest(data)
		assert.Equal(t, expected, result)
	})

	t.Run("non-card payment method", func(t *testing.T) {
		createdAt := time.Date(2026, 2, 3, 10, 30, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 2, 3, 10, 35, 0, 0, time.UTC)

		data := &fdscommon.CheckRequest{
			Partner: fdscommon.PartnerCheck{
				ID:        "merchant-456",                 // NOSONAR
				Company:   util.ValueToPtr("VA Merchant"), // NOSONAR
				RiskLevel: "LOW",                          // NOSONAR
			},
			Customer: fdscommon.CustomerCheck{
				FirstName: util.ValueToPtr("Jane"),             // NOSONAR
				Email:     util.ValueToPtr("jane@example.com"), // NOSONAR
				Phone:     util.ValueToPtr("+6289876543210"),   // NOSONAR
			},
			Transaction: fdscommon.TransactionCheck{
				ID:                "order-va-789",                               // NOSONAR
				ClientReferenceID: "va-ref-123",                                 // NOSONAR
				OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(50000)), // NOSONAR
				OrderCurrency:     util.ValueToPtr("IDR"),                       // NOSONAR
				CreatedAt:         createdAt,                                    // NOSONAR
				UpdatedAt:         updatedAt,                                    // NOSONAR
			},
			Payment: fdscommon.PaymentCheck{
				MethodType: "VIRTUAL_ACCOUNT", // NOSONAR
			},
			Device: fdscommon.DeviceCheck{
				IPType:    "v4",                        // NOSONAR
				IPAddress: util.ValueToPtr("10.0.0.1"), // NOSONAR
			},
		}

		expected := model.PaymentWorkflowRequest{
			Merchant: model.Merchant{
				ID:        "merchant-456", // NOSONAR
				Name:      "VA Merchant",  // NOSONAR
				RiskLevel: "LOW",          // NOSONAR
			},
			Customer: model.Customer{
				Name:        "Jane",             // NOSONAR
				Email:       "jane@example.com", // NOSONAR
				PhoneNumber: "+6289876543210",   // NOSONAR
			},
			Transaction: model.Transaction{
				ID:                "order-va-789", // NOSONAR
				ClientReferenceID: "va-ref-123",   // NOSONAR
				Amount: common.Amount2{
					Value:    50000, // NOSONAR
					Currency: "IDR", // NOSONAR
				},
				CreatedAt: createdAt, // NOSONAR
				UpdatedAt: updatedAt, // NOSONAR
			},
			PaymentMethod: model.PaymentMethod{
				Type: "VIRTUAL_ACCOUNT", // NOSONAR
			},
			Device: model.Device{
				IPType:    "v4",       // NOSONAR
				IPAddress: "10.0.0.1", // NOSONAR
			},
			Metadata: map[string]any{},
		}

		result := toPaymentWorkflowRequest(data)
		assert.Equal(t, expected, result)
	})

	t.Run("with nil optional fields", func(t *testing.T) {
		createdAt := time.Date(2026, 2, 3, 10, 30, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 2, 3, 10, 35, 0, 0, time.UTC)

		data := &fdscommon.CheckRequest{
			Partner: fdscommon.PartnerCheck{
				ID:        "merchant-789",                  // NOSONAR
				Company:   util.ValueToPtr("Test Company"), // NOSONAR
				RiskLevel: "HIGH",                          // NOSONAR
			},
			Customer: fdscommon.CustomerCheck{
				FirstName: nil,
				Email:     nil,
				Phone:     nil,
			},
			Transaction: fdscommon.TransactionCheck{
				ID:                "order-nil-123",                              // NOSONAR
				ClientReferenceID: "nil-ref-456",                                // NOSONAR
				OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(75000)), // NOSONAR
				OrderCurrency:     util.ValueToPtr("IDR"),                       // NOSONAR
				CreatedAt:         createdAt,                                    // NOSONAR
				UpdatedAt:         updatedAt,                                    // NOSONAR
			},
			Payment: fdscommon.PaymentCheck{
				MethodType: "QRIS", // NOSONAR
			},
			Device: fdscommon.DeviceCheck{
				IPAddress: nil,
			},
		}

		expected := model.PaymentWorkflowRequest{
			Merchant: model.Merchant{
				ID:        "merchant-789", // NOSONAR
				Name:      "Test Company", // NOSONAR
				RiskLevel: "HIGH",         // NOSONAR
			},
			Customer: model.Customer{
				Name:        "",
				Email:       "",
				PhoneNumber: "",
			},
			Transaction: model.Transaction{
				ID:                "order-nil-123", // NOSONAR
				ClientReferenceID: "nil-ref-456",   // NOSONAR
				Amount: common.Amount2{
					Value:    75000, // NOSONAR
					Currency: "IDR", // NOSONAR
				},
				CreatedAt: createdAt, // NOSONAR
				UpdatedAt: updatedAt, // NOSONAR
			},
			PaymentMethod: model.PaymentMethod{
				Type: "QR", // NOSONAR
			},
			Device: model.Device{
				IPType:    "",
				IPAddress: "",
			},
			Metadata: map[string]any{},
		}

		result := toPaymentWorkflowRequest(data)
		assert.Equal(t, expected, result)
	})

	t.Run("card payment without custom data", func(t *testing.T) {
		createdAt := time.Date(2026, 2, 3, 10, 30, 0, 0, time.UTC)
		updatedAt := time.Date(2026, 2, 3, 10, 35, 0, 0, time.UTC)

		data := &fdscommon.CheckRequest{
			Partner: fdscommon.PartnerCheck{
				ID:        "merchant-no-custom",               // NOSONAR
				Company:   util.ValueToPtr("Simple Merchant"), // NOSONAR
				RiskLevel: "LOW",                              // NOSONAR
			},
			Customer: fdscommon.CustomerCheck{
				FirstName: util.ValueToPtr("Alice"),             // NOSONAR
				Email:     util.ValueToPtr("alice@example.com"), // NOSONAR
				Phone:     util.ValueToPtr("+6281111111111"),    // NOSONAR
			},
			Transaction: fdscommon.TransactionCheck{
				ID:                "order-simple",                               // NOSONAR
				ClientReferenceID: "simple-ref",                                 // NOSONAR
				OrderTotal:        util.ValueToPtr(decimal.NewFromFloat(25000)), // NOSONAR
				OrderCurrency:     util.ValueToPtr("IDR"),                       // NOSONAR
				CreatedAt:         createdAt,                                    // NOSONAR
				UpdatedAt:         updatedAt,                                    // NOSONAR
			},
			Payment: fdscommon.PaymentCheck{
				MethodType:       "CREDIT_CARD",      // NOSONAR
				Fingerprint:      "fp-simple",        // NOSONAR
				MaskedCardNumber: "520000******0005", // NOSONAR
				CardBrand:        "MASTERCARD",       // NOSONAR
				CardCountryCode:  "US",               // NOSONAR
				CardType:         "DEBIT",            // NOSONAR
				CardIssuing:      "Bank XYZ",         // NOSONAR
				ThreeDsEci:       nil,
				AuthCode:         nil,
				CvvResultCode:    nil,
			},
			Device: fdscommon.DeviceCheck{
				IPType:    "v6",                            // NOSONAR
				IPAddress: util.ValueToPtr("2001:0db8::1"), // NOSONAR
			},
			Custom: nil,
		}

		expected := model.PaymentWorkflowRequest{
			Merchant: model.Merchant{
				ID:        "merchant-no-custom", // NOSONAR
				Name:      "Simple Merchant",    // NOSONAR
				RiskLevel: "LOW",                // NOSONAR
			},
			Customer: model.Customer{
				Name:        "Alice",             // NOSONAR
				Email:       "alice@example.com", // NOSONAR
				PhoneNumber: "+6281111111111",    // NOSONAR
			},
			Transaction: model.Transaction{
				ID:                "order-simple", // NOSONAR
				ClientReferenceID: "simple-ref",   // NOSONAR
				Amount: common.Amount2{
					Value:    25000, // NOSONAR
					Currency: "IDR", // NOSONAR
				},
				CreatedAt: createdAt, // NOSONAR
				UpdatedAt: updatedAt, // NOSONAR
			},
			PaymentMethod: model.PaymentMethod{
				Type: "CARD", // NOSONAR
				Card: &model.PaymentMethodTypeCard{
					CardFingerprint: "fp-simple",        // NOSONAR
					CardNumber:      "520000******0005", // NOSONAR
					CardBrand:       "MASTERCARD",       // NOSONAR
					CardCountryCode: "US",               // NOSONAR
					CardType:        "DEBIT",            // NOSONAR
					IssuerName:      "Bank XYZ",         // NOSONAR
					ECICode:         "",
					ApprovalCode:    "",
					CvvCode:         "",
					BankMerchantID:  "",
					AcquirerName:    "",
				},
			},
			Device: model.Device{
				IPType:    "v6",           // NOSONAR
				IPAddress: "2001:0db8::1", // NOSONAR
			},
			Metadata: map[string]any{},
		}

		result := toPaymentWorkflowRequest(data)
		assert.Equal(t, expected, result)
	})
}

func TestToPaymentMethodTypeCard(t *testing.T) {

	t.Run("with all fields populated including custom data", func(t *testing.T) {
		data := &fdscommon.CheckRequest{
			Payment: fdscommon.PaymentCheck{
				Fingerprint:      "fingerprint-abc123",       // NOSONAR
				MaskedCardNumber: "411111******1111",         // NOSONAR
				CardBrand:        "VISA",                     // NOSONAR
				CardCountryCode:  "ID",                       // NOSONAR
				CardType:         "CREDIT",                   // NOSONAR
				CardIssuing:      "Bank Indonesia",           // NOSONAR
				ThreeDsEci:       util.ValueToPtr("05"),      // NOSONAR
				AuthCode:         util.ValueToPtr("AUTH456"), // NOSONAR
				CvvResultCode:    util.ValueToPtr("M"),       // NOSONAR
			},
			Custom: &fdscommon.CustomCheck{
				Number:        util.ValueToPtr("MID-12345"),     // NOSONAR
				AcquiringName: util.ValueToPtr("Main Acquirer"), // NOSONAR
			},
		}

		expected := &model.PaymentMethodTypeCard{
			CardFingerprint: "fingerprint-abc123", // NOSONAR
			CardNumber:      "411111******1111",   // NOSONAR
			CardBrand:       "VISA",               // NOSONAR
			CardCountryCode: "ID",                 // NOSONAR
			CardType:        "CREDIT",             // NOSONAR
			IssuerName:      "Bank Indonesia",     // NOSONAR
			ECICode:         "05",                 // NOSONAR
			ApprovalCode:    "AUTH456",            // NOSONAR
			CvvCode:         "M",                  // NOSONAR
			BankMerchantID:  "MID-12345",          // NOSONAR
			AcquirerName:    "Main Acquirer",      // NOSONAR
		}

		result := toPaymentMethodTypeCard(data)
		assert.Equal(t, expected, result)
	})

	t.Run("with nil optional fields", func(t *testing.T) {
		data := &fdscommon.CheckRequest{
			Payment: fdscommon.PaymentCheck{
				Fingerprint:      "fingerprint-xyz789", // NOSONAR
				MaskedCardNumber: "520000******0005",   // NOSONAR
				CardBrand:        "MASTERCARD",         // NOSONAR
				CardCountryCode:  "US",                 // NOSONAR
				CardType:         "DEBIT",              // NOSONAR
				CardIssuing:      "Bank USA",           // NOSONAR
				ThreeDsEci:       nil,
				AuthCode:         nil,
				CvvResultCode:    nil,
			},
			Custom: &fdscommon.CustomCheck{
				Number:        util.ValueToPtr("MID-67890"),          // NOSONAR
				AcquiringName: util.ValueToPtr("Secondary Acquirer"), // NOSONAR
			},
		}

		expected := &model.PaymentMethodTypeCard{
			CardFingerprint: "fingerprint-xyz789", // NOSONAR
			CardNumber:      "520000******0005",   // NOSONAR
			CardBrand:       "MASTERCARD",         // NOSONAR
			CardCountryCode: "US",                 // NOSONAR
			CardType:        "DEBIT",              // NOSONAR
			IssuerName:      "Bank USA",           // NOSONAR
			ECICode:         "",
			ApprovalCode:    "",
			CvvCode:         "",
			BankMerchantID:  "MID-67890",          // NOSONAR
			AcquirerName:    "Secondary Acquirer", // NOSONAR
		}

		result := toPaymentMethodTypeCard(data)
		assert.Equal(t, expected, result)
	})

	t.Run("without custom data", func(t *testing.T) {
		data := &fdscommon.CheckRequest{
			Payment: fdscommon.PaymentCheck{
				Fingerprint:      "fingerprint-def456",       // NOSONAR
				MaskedCardNumber: "378282******0005",         // NOSONAR
				CardBrand:        "AMEX",                     // NOSONAR
				CardCountryCode:  "SG",                       // NOSONAR
				CardType:         "CREDIT",                   // NOSONAR
				CardIssuing:      "Bank Singapore",           // NOSONAR
				ThreeDsEci:       util.ValueToPtr("07"),      // NOSONAR
				AuthCode:         util.ValueToPtr("AUTH789"), // NOSONAR
				CvvResultCode:    util.ValueToPtr("P"),       // NOSONAR
			},
			Custom: nil,
		}

		expected := &model.PaymentMethodTypeCard{
			CardFingerprint: "fingerprint-def456", // NOSONAR
			CardNumber:      "378282******0005",   // NOSONAR
			CardBrand:       "AMEX",               // NOSONAR
			CardCountryCode: "SG",                 // NOSONAR
			CardType:        "CREDIT",             // NOSONAR
			IssuerName:      "Bank Singapore",     // NOSONAR
			ECICode:         "07",                 // NOSONAR
			ApprovalCode:    "AUTH789",            // NOSONAR
			CvvCode:         "P",                  // NOSONAR
			BankMerchantID:  "",
			AcquirerName:    "",
		}

		result := toPaymentMethodTypeCard(data)
		assert.Equal(t, expected, result)
	})

	t.Run("with empty custom fields", func(t *testing.T) {
		data := &fdscommon.CheckRequest{
			Payment: fdscommon.PaymentCheck{
				Fingerprint:      "fingerprint-ghi123",       // NOSONAR
				MaskedCardNumber: "601100******0001",         // NOSONAR
				CardBrand:        "DISCOVER",                 // NOSONAR
				CardCountryCode:  "AU",                       // NOSONAR
				CardType:         "PREPAID",                  // NOSONAR
				CardIssuing:      "Bank Australia",           // NOSONAR
				ThreeDsEci:       util.ValueToPtr("02"),      // NOSONAR
				AuthCode:         util.ValueToPtr("AUTH000"), // NOSONAR
				CvvResultCode:    util.ValueToPtr("N"),       // NOSONAR
			},
			Custom: &fdscommon.CustomCheck{
				Number:        nil,
				AcquiringName: nil,
			},
		}

		expected := &model.PaymentMethodTypeCard{
			CardFingerprint: "fingerprint-ghi123", // NOSONAR
			CardNumber:      "601100******0001",   // NOSONAR
			CardBrand:       "DISCOVER",           // NOSONAR
			CardCountryCode: "AU",                 // NOSONAR
			CardType:        "PREPAID",            // NOSONAR
			IssuerName:      "Bank Australia",     // NOSONAR
			ECICode:         "02",                 // NOSONAR
			ApprovalCode:    "AUTH000",            // NOSONAR
			CvvCode:         "N",                  // NOSONAR
			BankMerchantID:  "",
			AcquirerName:    "",
		}

		result := toPaymentMethodTypeCard(data)
		assert.Equal(t, expected, result)
	})
}
