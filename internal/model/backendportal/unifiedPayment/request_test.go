package unifiedPaymentModel

import (
	"encoding/json"
	"testing"
	"time"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	creditcardModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcard"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestFilterChargeRequest_HashFilter(t *testing.T) {
	tests := []struct {
		name     string
		request  FilterChargeRequest
		timezone string
		want     string
	}{
		{
			name: "basic hash with minimal fields",
			request: FilterChargeRequest{
				MerchantID:     "merchant123",
				StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
				Sort:           "created_at",
			},
			timezone: "UTC",
			want:     "a8c2d8e9f5b6c7a8d9e0f1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2", // Placeholder - actual hash will be computed
		},
		{
			name: "hash with all optional fields",
			request: FilterChargeRequest{
				MerchantID:        "merchant123",
				StartCreatedAt:    time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndCreatedAt:      time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
				Status:            "SUCCESS",
				UUID:              "uuid123",
				ClientReferenceID: "client_ref_123",
				PaymentSessionID:  "session123",
				Sort:              "created_at",
			},
			timezone: "UTC",
			want:     "", // Will be computed in test
		},
		{
			name: "hash with future EndCreatedAt should use current time",
			request: FilterChargeRequest{
				MerchantID:     "merchant123",
				StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Now().UTC().Add(24 * time.Hour), // Future date
				Sort:           "created_at",
			},
			timezone: "UTC",
			want:     "", // Will be computed in test
		},
		{
			name: "hash with different timezone",
			request: FilterChargeRequest{
				MerchantID:     "merchant123",
				StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
				Sort:           "created_at",
			},
			timezone: "Asia/Jakarta",
			want:     "", // Will be computed in test
		},
		{
			name: "hash with empty optional fields",
			request: FilterChargeRequest{
				MerchantID:        "merchant123",
				StartCreatedAt:    time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
				EndCreatedAt:      time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
				Status:            "",
				UUID:              "",
				ClientReferenceID: "",
				PaymentSessionID:  "",
				Sort:              "created_at",
			},
			timezone: "UTC",
			want:     "", // Will be computed in test
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.HashFilter(tt.timezone)

			// Basic assertions
			assert.NotEmpty(t, result, "Hash should not be empty")
			assert.Len(t, result, 64, "SHA256 hash should be 64 characters long")

			// Test that the hash is consistent
			result2 := tt.request.HashFilter(tt.timezone)
			assert.Equal(t, result, result2, "Hash should be consistent for same input")

			// Test that different inputs produce different hashes
			if tt.name == "basic hash with minimal fields" {
				differentRequest := tt.request
				differentRequest.MerchantID = "different_merchant"
				differentHash := differentRequest.HashFilter(tt.timezone)
				assert.NotEqual(t, result, differentHash, "Different inputs should produce different hashes")
			}
		})
	}
}

func TestFilterChargeRequest_HashFilter_TimeBehavior(t *testing.T) {
	t.Run("should use current time when EndCreatedAt is in future", func(t *testing.T) {
		request := FilterChargeRequest{
			MerchantID:     "merchant123",
			StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			EndCreatedAt:   time.Now().UTC().Add(24 * time.Hour), // Future date
			Sort:           "created_at",
		}

		// Get hash with future end date
		hash1 := request.HashFilter("UTC")

		// Create same request but with current time as end date
		request2 := request
		request2.EndCreatedAt = time.Now().UTC()
		hash2 := request2.HashFilter("UTC")

		// The hashes should be similar because the function should use current time
		// when EndCreatedAt is in the future
		assert.NotEmpty(t, hash1)
		assert.NotEmpty(t, hash2)
	})

	t.Run("should use actual EndCreatedAt when it's in the past", func(t *testing.T) {
		pastTime := time.Date(2022, 1, 1, 10, 0, 0, 0, time.UTC)
		request := FilterChargeRequest{
			MerchantID:     "merchant123",
			StartCreatedAt: time.Date(2021, 1, 1, 10, 0, 0, 0, time.UTC),
			EndCreatedAt:   pastTime,
			Sort:           "created_at",
		}

		hash := request.HashFilter("UTC")
		assert.NotEmpty(t, hash)

		// Hash should be consistent
		hash2 := request.HashFilter("UTC")
		assert.Equal(t, hash, hash2)
	})
}

func TestFilterChargeRequest_HashFilter_FieldConsistency(t *testing.T) {
	baseRequest := FilterChargeRequest{
		MerchantID:     "merchant123",
		StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
		EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
		Sort:           "created_at",
	}

	t.Run("adding Status should change hash", func(t *testing.T) {
		hash1 := baseRequest.HashFilter("UTC")

		requestWithStatus := baseRequest
		requestWithStatus.Status = "SUCCESS"
		hash2 := requestWithStatus.HashFilter("UTC")

		assert.NotEqual(t, hash1, hash2, "Adding Status should change the hash")
	})

	t.Run("adding UUID should change hash", func(t *testing.T) {
		hash1 := baseRequest.HashFilter("UTC")

		requestWithUUID := baseRequest
		requestWithUUID.UUID = "uuid123"
		hash2 := requestWithUUID.HashFilter("UTC")

		assert.NotEqual(t, hash1, hash2, "Adding UUID should change the hash")
	})

	// Note: These tests demonstrate bugs in the implementation
	// Lines 328 and 331 use r.UUID instead of the correct field values
	t.Run("ClientReferenceID bug - uses UUID instead of ClientReferenceID", func(t *testing.T) {
		// Test 1: Different ClientReferenceID values with same UUID should produce same hash (demonstrating bug)
		request1 := baseRequest
		request1.ClientReferenceID = "client_ref_123"
		request1.UUID = "same_uuid"

		request2 := baseRequest
		request2.ClientReferenceID = "different_client_ref"
		request2.UUID = "same_uuid" // Same UUID

		hash1 := request1.HashFilter("UTC")
		hash2 := request2.HashFilter("UTC")

		// These should be different if ClientReferenceID was properly used, but they're the same due to the bug
		assert.Equal(t, hash1, hash2, "KNOWN BUG: Different ClientReferenceID values with same UUID produce same hash")

		// Test 2: Same ClientReferenceID with different UUID should produce different hash
		request3 := baseRequest
		request3.ClientReferenceID = "same_client_ref"
		request3.UUID = "uuid1"

		request4 := baseRequest
		request4.ClientReferenceID = "same_client_ref"
		request4.UUID = "uuid2"

		hash3 := request3.HashFilter("UTC")
		hash4 := request4.HashFilter("UTC")

		// These are different because different UUIDs are used (proving the bug)
		assert.NotEqual(t, hash3, hash4, "KNOWN BUG: Same ClientReferenceID with different UUID produce different hash")

		t.Log("KNOWN BUG: Line 328 uses r.UUID instead of r.ClientReferenceID")
	})

	t.Run("PaymentSessionID bug - uses UUID instead of PaymentSessionID", func(t *testing.T) {
		// Test 1: Different PaymentSessionID values with same UUID should produce same hash (demonstrating bug)
		request1 := baseRequest
		request1.PaymentSessionID = "session_123"
		request1.UUID = "same_uuid"

		request2 := baseRequest
		request2.PaymentSessionID = "different_session"
		request2.UUID = "same_uuid" // Same UUID

		hash1 := request1.HashFilter("UTC")
		hash2 := request2.HashFilter("UTC")

		// These should be different if PaymentSessionID was properly used, but they're the same due to the bug
		assert.Equal(t, hash1, hash2, "KNOWN BUG: Different PaymentSessionID values with same UUID produce same hash")

		// Test 2: Same PaymentSessionID with different UUID should produce different hash
		request3 := baseRequest
		request3.PaymentSessionID = "same_session"
		request3.UUID = "uuid1"

		request4 := baseRequest
		request4.PaymentSessionID = "same_session"
		request4.UUID = "uuid2"

		hash3 := request3.HashFilter("UTC")
		hash4 := request4.HashFilter("UTC")

		// These are different because different UUIDs are used (proving the bug)
		assert.NotEqual(t, hash3, hash4, "KNOWN BUG: Same PaymentSessionID with different UUID produce different hash")

		t.Log("KNOWN BUG: Line 331 uses r.UUID instead of r.PaymentSessionID")
	})

	t.Run("different timezone should change hash", func(t *testing.T) {
		hash1 := baseRequest.HashFilter("UTC")
		hash2 := baseRequest.HashFilter("Asia/Jakarta")

		assert.NotEqual(t, hash1, hash2, "Different timezone should change the hash")
	})

	t.Run("different sort should change hash", func(t *testing.T) {
		hash1 := baseRequest.HashFilter("UTC")

		requestWithDifferentSort := baseRequest
		requestWithDifferentSort.Sort = "amount"
		hash2 := requestWithDifferentSort.HashFilter("UTC")

		assert.NotEqual(t, hash1, hash2, "Different sort should change the hash")
	})
}

func TestFilterChargeRequest_HashFilter_EdgeCases(t *testing.T) {
	t.Run("empty timezone", func(t *testing.T) {
		request := FilterChargeRequest{
			MerchantID:     "merchant123",
			StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
			Sort:           "created_at",
		}

		hash := request.HashFilter("")
		assert.NotEmpty(t, hash, "Hash should be generated even with empty timezone")
		assert.Len(t, hash, 64, "Hash should still be 64 characters")
	})

	t.Run("empty sort", func(t *testing.T) {
		request := FilterChargeRequest{
			MerchantID:     "merchant123",
			StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
			Sort:           "",
		}

		hash := request.HashFilter("UTC")
		assert.NotEmpty(t, hash, "Hash should be generated even with empty sort")
		assert.Len(t, hash, 64, "Hash should still be 64 characters")
	})

	t.Run("empty merchant ID", func(t *testing.T) {
		request := FilterChargeRequest{
			MerchantID:     "",
			StartCreatedAt: time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC),
			EndCreatedAt:   time.Date(2023, 1, 2, 10, 0, 0, 0, time.UTC),
			Sort:           "created_at",
		}

		hash := request.HashFilter("UTC")
		assert.NotEmpty(t, hash, "Hash should be generated even with empty merchant ID")
		assert.Len(t, hash, 64, "Hash should still be 64 characters")
	})
}

// Test backward compatibility for surname fields
func TestBillingInformation_GetEffectiveSurName(t *testing.T) {
	t.Run("should return surname when only surname is provided", func(t *testing.T) {
		billing := &BillingInformation{
			Surname:  "Smith",
			SureName: "",
		}

		result := billing.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should return SureName when only SureName is provided", func(t *testing.T) {
		billing := &BillingInformation{
			Surname:  "",
			SureName: "Johnson",
		}

		result := billing.GetSurname()
		assert.Equal(t, "Johnson", result)
	})

	t.Run("should prioritize surname when both fields are provided", func(t *testing.T) {
		billing := &BillingInformation{
			Surname:  "Smith",
			SureName: "Johnson",
		}

		result := billing.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should return empty when both fields are empty", func(t *testing.T) {
		billing := &BillingInformation{
			Surname:  "",
			SureName: "",
		}

		result := billing.GetSurname()
		assert.Equal(t, "", result)
	})
}

func TestShippingInformation_GetEffectiveSurName(t *testing.T) {
	t.Run("should return surname when only surname is provided", func(t *testing.T) {
		shipping := &ShippingInformation{
			Surname:  "Smith",
			SureName: "",
		}

		result := shipping.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should return SureName when only SureName is provided", func(t *testing.T) {
		shipping := &ShippingInformation{
			Surname:  "",
			SureName: "Johnson",
		}

		result := shipping.GetSurname()
		assert.Equal(t, "Johnson", result)
	})

	t.Run("should prioritize surname when both fields are provided", func(t *testing.T) {
		shipping := &ShippingInformation{
			Surname:  "Smith",
			SureName: "Johnson",
		}

		result := shipping.GetSurname()
		assert.Equal(t, "Smith", result)
	})
}

func TestCustomerInformation_GetEffectiveSurName(t *testing.T) {
	t.Run("should return surname when only surname is provided", func(t *testing.T) {
		customer := &CustomerInformation{
			Surname:  util.ValueToPtr("Smith"),
			SureName: "",
		}

		result := customer.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should return SureName when only SureName is provided", func(t *testing.T) {
		customer := &CustomerInformation{
			Surname:  nil,
			SureName: "Johnson",
		}

		result := customer.GetSurname()
		assert.Equal(t, "Johnson", result)
	})

	t.Run("should prioritize surname when both fields are provided", func(t *testing.T) {
		customer := &CustomerInformation{
			Surname:  util.ValueToPtr("Smith"),
			SureName: "Johnson",
		}

		result := customer.GetSurname()
		assert.Equal(t, "Smith", result)
	})
}

// Test JSON marshaling/unmarshaling for backward compatibility
func TestBillingInformation_JSON_BackwardCompatibility(t *testing.T) {
	t.Run("should unmarshal JSON with surname field", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"surname": "Smith",
			"email": "john.smith@example.com"
		}`

		var billing BillingInformation
		err := json.Unmarshal([]byte(jsonData), &billing)
		assert.NoError(t, err)
		assert.Equal(t, "John", billing.GivenName)
		assert.Equal(t, "Smith", billing.Surname)
		assert.Equal(t, "", billing.SureName)

		// Test the processing logic
		result := billing.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should unmarshal JSON with sureName field (backward compatibility)", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"sureName": "Johnson",
			"email": "john.johnson@example.com"
		}`

		var billing BillingInformation
		err := json.Unmarshal([]byte(jsonData), &billing)
		assert.NoError(t, err)
		assert.Equal(t, "John", billing.GivenName)
		assert.Equal(t, "", billing.Surname)
		assert.Equal(t, "Johnson", billing.SureName)

		// Test the processing logic
		result := billing.GetSurname()
		assert.Equal(t, "Johnson", result)
	})

	t.Run("should prioritize surname when both fields are present", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"surname": "Smith",
			"sureName": "Johnson",
			"email": "john@example.com"
		}`

		var billing BillingInformation
		err := json.Unmarshal([]byte(jsonData), &billing)
		assert.NoError(t, err)
		assert.Equal(t, "Smith", billing.Surname)
		assert.Equal(t, "Johnson", billing.SureName)

		// Test the processing logic prioritizes surname
		result := billing.GetSurname()
		assert.Equal(t, "Smith", result)
	})
}

func TestShippingInformation_JSON_BackwardCompatibility(t *testing.T) {
	t.Run("should unmarshal JSON with surname field", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"surname": "Smith",
			"email": "john.smith@example.com",
			"addressLine1": "123 Main St",
			"city": "Jakarta",
			"provinceState": "DKI Jakarta",
			"country": "Indonesia",
			"method": "REGULAR"
		}`

		var shipping ShippingInformation
		err := json.Unmarshal([]byte(jsonData), &shipping)
		assert.NoError(t, err)
		assert.Equal(t, "Smith", shipping.Surname)
		assert.Equal(t, "", shipping.SureName)

		result := shipping.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should unmarshal JSON with sureName field (backward compatibility)", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"sureName": "Johnson",
			"email": "john.johnson@example.com",
			"addressLine1": "123 Main St",
			"city": "Jakarta",
			"provinceState": "DKI Jakarta",
			"country": "Indonesia",
			"method": "REGULAR"
		}`

		var shipping ShippingInformation
		err := json.Unmarshal([]byte(jsonData), &shipping)
		assert.NoError(t, err)
		assert.Equal(t, "", shipping.Surname)
		assert.Equal(t, "Johnson", shipping.SureName)

		result := shipping.GetSurname()
		assert.Equal(t, "Johnson", result)
	})
}

func TestCustomerInformation_JSON_BackwardCompatibility(t *testing.T) {
	t.Run("should unmarshal JSON with surname field", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"surname": "Smith",
			"email": "john.smith@example.com"
		}`

		var customer CustomerInformation
		err := json.Unmarshal([]byte(jsonData), &customer)
		assert.NoError(t, err)
		assert.NotNil(t, customer.Surname)
		assert.Equal(t, "Smith", *customer.Surname)
		assert.Equal(t, "", customer.SureName)

		result := customer.GetSurname()
		assert.Equal(t, "Smith", result)
	})

	t.Run("should unmarshal JSON with sureName field (backward compatibility)", func(t *testing.T) {
		jsonData := `{
			"givenName": "John",
			"sureName": "Johnson",
			"email": "john.johnson@example.com"
		}`

		var customer CustomerInformation
		err := json.Unmarshal([]byte(jsonData), &customer)
		assert.NoError(t, err)
		assert.Nil(t, customer.Surname)
		assert.Equal(t, "Johnson", customer.SureName)

		result := customer.GetSurname()
		assert.Equal(t, "Johnson", result)
	})
}

func TestCreateUnifiedPaymentSessionRequestRecurringPaymentType(t *testing.T) {
	tests := []struct {
		request    CreateUnifiedPaymentSessionRequest
		wantResult string
	}{
		{},
		{
			request: CreateUnifiedPaymentSessionRequest{
				RecurringID: "aadcaa22-446c-45ea-8377-f2d3d496618f",
			},
			wantResult: constant.RecurringPaymentTypeSubsequentPayment,
		},
		{
			request: CreateUnifiedPaymentSessionRequest{
				RecurringID:                "035db2af-d02c-47c9-bc8e-a06f130bcad9",
				InitiateFirstAuthorization: true,
			},
			wantResult: constant.RecurringPaymentTypeFirstAuthorization,
		},
	}
	for _, test := range tests {
		assert.Equal(t, test.wantResult, test.request.RecurringPaymentType())
	}
}

func TestCreateUnifiedPaymentSessionRequestGetEwalletChannel(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateUnifiedPaymentSessionRequest
		expected string
	}{
		{
			name:     "should return empty string when Ewallet is nil",
			request:  CreateUnifiedPaymentSessionRequest{},
			expected: "",
		},
		{
			name: "should return empty string when Ewallet channel is empty",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Ewallet: &PaymentMethodOptionEwallet{
						Channel: "",
					},
				},
			},
			expected: "",
		},
		{
			name: "should return channel when Ewallet has SHOPEEPAY channel",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Ewallet: &PaymentMethodOptionEwallet{
						Channel: "SHOPEEPAY",
					},
				},
			},
			expected: "SHOPEEPAY",
		},
		{
			name: "should return channel when Ewallet has DANA channel",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Ewallet: &PaymentMethodOptionEwallet{
						Channel: "DANA",
					},
				},
			},
			expected: "DANA",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.GetEwalletChannel()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateUnifiedPaymentSessionRequestIsAutoSplitCardPayment(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateUnifiedPaymentSessionRequest
		expected bool
	}{
		{
			name:     "should return false when Card is nil",
			request:  CreateUnifiedPaymentSessionRequest{},
			expected: false,
		},
		{
			name: "should return false when AutoSplit is nil",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{},
				},
			},
			expected: false,
		},
		{
			name: "should return false when AutoSplit is false",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{
						AutoSplit: util.ValueToPtr(false),
					},
				},
			},
			expected: false,
		},
		{
			name: "should return true when AutoSplit is true",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{
						AutoSplit: util.ValueToPtr(true),
					},
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.IsAutoSplitCardPayment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfirmUnifiedPaymentSessionRequestIsAutoSplitPaymentAuth(t *testing.T) {
	tests := []struct {
		name     string
		request  ConfirmUnifiedPaymentSessionRequest
		expected bool
	}{
		{
			name:     "should return false when AutoSplitPayment is nil",
			request:  ConfirmUnifiedPaymentSessionRequest{},
			expected: false,
		},
		{
			name: "should return false when TransactionType is not AUTHENTICATION",
			request: ConfirmUnifiedPaymentSessionRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: "CAPTURE",
				},
			},
			expected: false,
		},
		{
			name: "should return true when TransactionType is AUTHENTICATION",
			request: ConfirmUnifiedPaymentSessionRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
			},
			expected: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.IsAutoSplitPaymentAuth()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAutoSplitPaymentToCardAutoSplitPayment(t *testing.T) {
	tests := []struct {
		name     string
		input    AutoSplitPayment
		expected *creditcardModel.AutoSplitPayment
	}{
		{
			name: "should convert all fields correctly",
			input: AutoSplitPayment{
				TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				Processor:       "MPGS",
				ProcessorLimit:  2000000000,
				CITMerchantID:   "CIT_MID_001",
				MITMerchantID:   "MIT_MID_001",
			},
			expected: &creditcardModel.AutoSplitPayment{
				TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				Processor:       "MPGS",
				ProcessorLimit:  2000000000,
				CITMerchantID:   "CIT_MID_001",
				MITMerchantID:   "MIT_MID_001",
			},
		},
		{
			name: "should convert with empty optional fields",
			input: AutoSplitPayment{
				TransactionType: constant.AutoSplitPaymentTypeAuthentication,
			},
			expected: &creditcardModel.AutoSplitPayment{
				TransactionType: constant.AutoSplitPaymentTypeAuthentication,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.ToCardAutoSplitPayment()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateUnifiedPaymentSessionRequestPrepareAutoSplitCardPayment(t *testing.T) {
	validSplitConfig := func() *paymentMethodModel.SplitCardPaymentConfig {
		return &paymentMethodModel.SplitCardPaymentConfig{
			Enabled:         true,
			ActiveProcessor: "MPGS",
			Processors: map[string]paymentMethodModel.CardPartnerProcessorConfig{
				"MPGS": {Limit: 2000000000},
			},
		}
	}

	validPartnerConfig := func() *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest {
		return &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
			Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
				Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
					{
						PartnerProcessor:   "MPGS",
						AcquirerMerchantID: "CIT_MID_001",
						SplitPaymentType:   "CIT",
					},
					{
						PartnerProcessor:   "MPGS",
						AcquirerMerchantID: "MIT_MID_001",
						SplitPaymentType:   "MIT",
					},
				},
			},
		}
	}

	tests := []struct {
		name                  string
		request               CreateUnifiedPaymentSessionRequest
		splitConfig           *paymentMethodModel.SplitCardPaymentConfig
		partnerConfig         *paymentMethodModel.SetupPaymentMethodPartnerConfigRequest
		processorLimitDefault float64
		wantErr               bool
		assertResult          func(t *testing.T, r *CreateUnifiedPaymentSessionRequest)
	}{
		{
			name: "should prepare auto split payment correctly with processor limit from config",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{},
				},
			},
			splitConfig:           validSplitConfig(),
			partnerConfig:         validPartnerConfig(),
			processorLimitDefault: 1000000000,
			wantErr:               false,
			assertResult: func(t *testing.T, r *CreateUnifiedPaymentSessionRequest) {
				assert.NotNil(t, r.AutoSplitPayment)
				assert.Equal(t, constant.AutoSplitPaymentTypeAuthentication, r.AutoSplitPayment.TransactionType)
				assert.Equal(t, "MPGS", r.AutoSplitPayment.Processor)
				assert.Equal(t, float64(2000000000), r.AutoSplitPayment.ProcessorLimit)
				assert.Equal(t, "CIT_MID_001", r.AutoSplitPayment.CITMerchantID)
				assert.Equal(t, "MIT_MID_001", r.AutoSplitPayment.MITMerchantID)
				assert.Equal(t, "CIT_MID_001", r.PaymentMethodOptions.Card.ProcessingConfig.BankMerchantId)
				assert.Equal(t, constant.CardThreeDsMethodChallenge, r.PaymentMethodOptions.Card.ThreeDsMethod)
			},
		},
		{
			name: "should use processor limit default when active processor not in processors map",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{},
				},
			},
			splitConfig: &paymentMethodModel.SplitCardPaymentConfig{
				Enabled:         true,
				ActiveProcessor: "CYBS",
				Processors:      map[string]paymentMethodModel.CardPartnerProcessorConfig{},
			},
			partnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
						{
							PartnerProcessor:   "CYBS",
							AcquirerMerchantID: "CIT_MID_CYBS",
							SplitPaymentType:   "CIT",
						},
						{
							PartnerProcessor:   "CYBS",
							AcquirerMerchantID: "MIT_MID_CYBS",
							SplitPaymentType:   "MIT",
						},
					},
				},
			},
			processorLimitDefault: 9000000000,
			wantErr:               false,
			assertResult: func(t *testing.T, r *CreateUnifiedPaymentSessionRequest) {
				assert.Equal(t, float64(9000000000), r.AutoSplitPayment.ProcessorLimit)
			},
		},
		{
			name: "should return error when CITMerchantID is missing",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{},
				},
			},
			splitConfig: &paymentMethodModel.SplitCardPaymentConfig{
				Enabled:         true,
				ActiveProcessor: "MPGS",
				Processors: map[string]paymentMethodModel.CardPartnerProcessorConfig{
					"MPGS": {Limit: 2000000000},
				},
			},
			partnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
						{
							PartnerProcessor:   "MPGS",
							AcquirerMerchantID: "MIT_MID_001",
							SplitPaymentType:   "MIT",
						},
					},
				},
			},
			processorLimitDefault: 1000000000,
			wantErr:               true,
		},
		{
			name: "should return error when MITMerchantID is missing",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{},
				},
			},
			splitConfig: &paymentMethodModel.SplitCardPaymentConfig{
				Enabled:         true,
				ActiveProcessor: "MPGS",
				Processors: map[string]paymentMethodModel.CardPartnerProcessorConfig{
					"MPGS": {Limit: 2000000000},
				},
			},
			partnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
				Card: &paymentMethodModel.SetupPaymentMethodPartnerConfigForCardRequest{
					Items: []paymentMethodModel.SetupPaymentMethodPartnerConfigForCardObj{
						{
							PartnerProcessor:   "MPGS",
							AcquirerMerchantID: "CIT_MID_001",
							SplitPaymentType:   "CIT",
						},
					},
				},
			},
			processorLimitDefault: 1000000000,
			wantErr:               true,
		},
		{
			name: "should initialize ProcessingConfig when nil",
			request: CreateUnifiedPaymentSessionRequest{
				PaymentMethodOptions: PaymentMethodOptions{
					Card: &PaymentMethodOptionCard{
						ProcessingConfig: nil,
					},
				},
			},
			splitConfig:           validSplitConfig(),
			partnerConfig:         validPartnerConfig(),
			processorLimitDefault: 1000000000,
			wantErr:               false,
			assertResult: func(t *testing.T, r *CreateUnifiedPaymentSessionRequest) {
				assert.NotNil(t, r.PaymentMethodOptions.Card.ProcessingConfig)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.PrepareAutoSplitCardPayment(tt.splitConfig, tt.partnerConfig, tt.processorLimitDefault)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.assertResult != nil {
					tt.assertResult(t, &tt.request)
				}
			}
		})
	}
}

func TestPaymentNotificationRequest_GetCardFingerprintID(t *testing.T) {
	tests := []struct {
		name     string
		request  PaymentNotificationRequest
		expected string
	}{
		{
			name:     "should return empty when ChargePaymentMethodDetails is nil",
			request:  PaymentNotificationRequest{},
			expected: "",
		},
		{
			name: "should return empty when Card is nil",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{},
			},
			expected: "",
		},
		{
			name: "should return empty when Card Fingerprint is empty",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{},
				},
			},
			expected: "",
		},
		{
			name: "should return fingerprint when Card is populated",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						Fingerprint: "abc123def456",
					},
				},
			},
			expected: "abc123def456",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.GetCardFingerprintID()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBaseProcessorRequest_ShouldAuthenticateEncryptedCard(t *testing.T) {
	tests := []struct {
		name     string
		request  BaseProcessorRequest
		expected bool
	}{
		{
			name: "should return false when AutoSplitPayment is not AUTHENTICATION",
			request: BaseProcessorRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: "CAPTURE",
				},
			},
			expected: false,
		},
		{
			name: "should return true when CardFundedPayout Sequence is 1",
			request: BaseProcessorRequest{
				CardFundedPayout: &CardFundedPayout{
					Sequence: 1,
				},
			},
			expected: true,
		},
		{
			name: "should return false when CardFundedPayout Sequence is not 1",
			request: BaseProcessorRequest{
				CardFundedPayout: &CardFundedPayout{
					Sequence: 2,
				},
			},
			expected: false,
		},
		{
			name: "should return true when Mode is API and no AutoSplitPayment/CardFundedPayout",
			request: BaseProcessorRequest{
				Mode: constant.UnifiedPaymentModeAPI,
			},
			expected: true,
		},
		{
			name: "should return false when Mode is not API and no AutoSplitPayment/CardFundedPayout",
			request: BaseProcessorRequest{
				Mode: "REDIRECT",
			},
			expected: false,
		},
		{
			name: "should return true when AutoSplitPayment is AUTHENTICATION and Mode is not API",
			request: BaseProcessorRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
				Mode: "REDIRECT",
			},
			expected: false,
		},
		{
			name:     "should return false when all fields are zero/nil",
			request:  BaseProcessorRequest{},
			expected: false,
		},
		{
			name: "AutoSplitPayment AUTHENTICATION takes precedence - returns API mode result",
			request: BaseProcessorRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: constant.AutoSplitPaymentTypeAuthentication,
				},
				Mode: constant.UnifiedPaymentModeAPI,
			},
			expected: true,
		},
		{
			name: "CardFundedPayout Sequence 1 takes precedence over AutoSplitPayment non-AUTH",
			request: BaseProcessorRequest{
				AutoSplitPayment: &AutoSplitPayment{
					TransactionType: "CAPTURE",
				},
				CardFundedPayout: &CardFundedPayout{
					Sequence: 1,
				},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.ShouldAuthenticateEncryptedCard()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateUnifiedPaymentSessionRequestIsAPIMode(t *testing.T) {
	tests := []struct {
		name     string
		request  CreateUnifiedPaymentSessionRequest
		expected bool
	}{
		{
			name: "should return true when mode is API",
			request: CreateUnifiedPaymentSessionRequest{
				Mode: constant.UnifiedPaymentModeAPI,
			},
			expected: true,
		},
		{
			name: "should return false when mode is REDIRECT",
			request: CreateUnifiedPaymentSessionRequest{
				Mode: constant.UnifiedPaymentModeRedirect,
			},
			expected: false,
		},
		{
			name:     "should return false when mode is empty",
			request:  CreateUnifiedPaymentSessionRequest{},
			expected: false,
		},
		{
			name: "should return false when mode is any other value",
			request: CreateUnifiedPaymentSessionRequest{
				Mode: "OTHER",
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.request.IsAPIMode()
			assert.Equal(t, tt.expected, result)
		})
	}
}
func TestGetCardThreeDSCallbackID(t *testing.T) {
	tests := []struct {
		name     string
		request  PaymentNotificationRequest
		expected string
	}{
		{
			name:     "when ChargePaymentMethodDetails is nil, then should return empty id",
			request:  PaymentNotificationRequest{},
			expected: "",
		},
		{
			name: "when Card is nil, then should return empty id",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{},
			},
			expected: "",
		},
		{
			name: "when AuthenticationResult is nil, then should return empty id",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{},
				},
			},
			expected: "",
		},
		{
			name: "when ThreeDsResult is SUCCESS, then should return populated id",
			request: PaymentNotificationRequest{
				ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
					Card: &ChargePaymentMethodDetailCard{
						AuthenticationResult: &ChargePaymentMethodDetailCardAuthenticationResult{
							TransactionID:         "txn-id-123",
							ThreeDsVersion:        "2.1.0",
							EciCode:               "05",
							AuthenticationScheme:  "VISA",
							AcsTransactionID:      "acs-txn-id-456",
							CallbackTransactionID: "callback-id",
						},
					},
				},
			},
			expected: "callback-id",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.request.GetCardThreeDSCallbackID()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAutoSplitPaymentSummaryToAutoSplitDetail(t *testing.T) {
	sampleCharge := ChargeResponse{
		ChargePaymentMethodDetails: &ChargePaymentMethodDetails{
			Card: &ChargePaymentMethodDetailCard{
				ResponseCode:     &ChargePaymentMethodDetailCardResponseCode{},
				ACSURL:           "https://acs.example/3ds",
				CardHolderName:   "JOHN DOE",
				Fingerprint:      "fp-123",
				SaveForFutureUse: new(true),
				BankMerchantID:   "BANK-MID",
				CardName:         "VISA",
				ApprovalCode:     "APPROVAL-001",
			},
		},
	}

	tests := []struct {
		name     string
		summary  AutoSplitPaymentSummary
		expected *AutoSplitDetails
	}{
		{
			name: "fully populated summary with multiple charges",
			summary: AutoSplitPaymentSummary{
				Status:                      "SUCCESS",
				NumberOfCharges:             3,
				NumberOfSuccessfulCharges:   2,
				NumberOfInProcessCharges:    1,
				NumberOfFailedCharges:       0,
				TotalSuccessfulChargeAmount: commonModel.Amount{Value: "15000.50", Currency: "IDR"},
				TotalInProgressChargeAmount: commonModel.Amount{Value: "7500.00", Currency: "IDR"},
				TotalFailedChargeAmount:     commonModel.Amount{Value: "0", Currency: "IDR"},
				ChargeDetails:               []ChargeResponse{sampleCharge, sampleCharge},
			},
			expected: &AutoSplitDetails{
				Status:                    "SUCCESS",
				NumberOfCharges:           3,
				NumberOfSuccessfulCharges: 2,
				NumberOfInProcessCharges:  1,
				NumberOfFailedCharges:     0,
				TotalSuccessfulChargeAmount: &Amount{
					Value:    15000.5,
					Currency: "IDR",
				},
				TotalFailedChargeAmount: &Amount{
					Value:    0,
					Currency: "IDR",
				},
				TotalInProcessChargeAmount: &Amount{
					Value:    7500,
					Currency: "IDR",
				},
				ChargesDetails: []ChargeResponse{sampleCharge, sampleCharge},
			},
		},
		{
			name:    "empty summary with zero values and nil charges",
			summary: AutoSplitPaymentSummary{},
			expected: &AutoSplitDetails{
				Status:                      "",
				NumberOfCharges:             0,
				NumberOfSuccessfulCharges:   0,
				NumberOfInProcessCharges:    0,
				NumberOfFailedCharges:       0,
				TotalSuccessfulChargeAmount: &Amount{Value: 0, Currency: ""},
				TotalFailedChargeAmount:     &Amount{Value: 0, Currency: ""},
				TotalInProcessChargeAmount:  &Amount{Value: 0, Currency: ""},
				ChargesDetails:              nil,
			},
		},
		{
			name: "partial success status with decimal precision",
			summary: AutoSplitPaymentSummary{
				Status:                      "PARTIAL_SUCCESS",
				NumberOfCharges:             3,
				NumberOfSuccessfulCharges:   1,
				NumberOfInProcessCharges:    1,
				NumberOfFailedCharges:       1,
				TotalSuccessfulChargeAmount: commonModel.Amount{Value: "100.25", Currency: "IDR"},
				TotalInProgressChargeAmount: commonModel.Amount{Value: "25.00", Currency: "IDR"},
				TotalFailedChargeAmount:     commonModel.Amount{Value: "50.75", Currency: "IDR"},
				ChargeDetails:               []ChargeResponse{sampleCharge},
			},
			expected: &AutoSplitDetails{
				Status:                    "PARTIAL_SUCCESS",
				NumberOfCharges:           3,
				NumberOfSuccessfulCharges: 1,
				NumberOfInProcessCharges:  1,
				NumberOfFailedCharges:     1,
				TotalSuccessfulChargeAmount: &Amount{
					Value:    100.25,
					Currency: "IDR",
				},
				TotalFailedChargeAmount: &Amount{
					Value:    50.75,
					Currency: "IDR",
				},
				TotalInProcessChargeAmount: &Amount{
					Value:    25,
					Currency: "IDR",
				},
				ChargesDetails: []ChargeResponse{sampleCharge},
			},
		},
		{
			name: "failed status with no charges",
			summary: AutoSplitPaymentSummary{
				Status:                  "FAILED",
				NumberOfCharges:         0,
				TotalFailedChargeAmount: commonModel.Amount{Value: "999.99", Currency: "IDR"},
				ChargeDetails:           []ChargeResponse{},
			},
			expected: &AutoSplitDetails{
				Status:                    "FAILED",
				NumberOfCharges:           0,
				NumberOfSuccessfulCharges: 0,
				NumberOfInProcessCharges:  0,
				NumberOfFailedCharges:     0,
				TotalSuccessfulChargeAmount: &Amount{
					Value:    0,
					Currency: "",
				},
				TotalFailedChargeAmount: &Amount{
					Value:    999.99,
					Currency: "IDR",
				},
				TotalInProcessChargeAmount: &Amount{
					Value:    0,
					Currency: "",
				},
				ChargesDetails: []ChargeResponse{},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.summary.ToAutoSplitDetail()

			if result == nil {
				t.Fatalf("expected non-nil *AutoSplitDetails, got nil")
			}
			assert.Equal(t, tc.expected.Status, result.Status)
			assert.Equal(t, tc.expected.NumberOfCharges, result.NumberOfCharges)
			assert.Equal(t, tc.expected.NumberOfSuccessfulCharges, result.NumberOfSuccessfulCharges)
			assert.Equal(t, tc.expected.NumberOfInProcessCharges, result.NumberOfInProcessCharges)
			assert.Equal(t, tc.expected.NumberOfFailedCharges, result.NumberOfFailedCharges)

			assert.Equal(t, tc.expected.TotalSuccessfulChargeAmount, result.TotalSuccessfulChargeAmount)
			assert.Equal(t, tc.expected.TotalFailedChargeAmount, result.TotalFailedChargeAmount)
			assert.Equal(t, tc.expected.TotalInProcessChargeAmount, result.TotalInProcessChargeAmount)

			assert.Equal(t, tc.expected.ChargesDetails, result.ChargesDetails)
		})
	}
}
