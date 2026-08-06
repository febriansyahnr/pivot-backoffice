package fraudrulesmodel

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestFraudRulesQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    FraudRulesQuery
		expected string
	}{
		{
			name:     "all fields empty",
			query:    FraudRulesQuery{},
			expected: "",
		},
		{
			name: "only uuid provided",
			query: FraudRulesQuery{
				UUID: "1234",
			},
			expected: "uuid = '1234'",
		},
		{
			name: "only rule name provided",
			query: FraudRulesQuery{
				RuleName: "test_rule",
			},
			expected: "name like '%test_rule%'",
		},
		{
			name: "only reference type provided",
			query: FraudRulesQuery{
				ReferenceType: "BANK",
			},
			expected: `(JSON_CONTAINS(reference_type, '"BANK"') OR JSON_CONTAINS(reference_type, '"ANY"'))`,
		},
		{
			name: "uuid and rule name provided",
			query: FraudRulesQuery{
				UUID:     "5678",
				RuleName: "fraud_check",
			},
			expected: "uuid = '5678' AND name like '%fraud_check%'",
		},
		{
			name: "all fields provided",
			query: FraudRulesQuery{
				UUID:          "abcd",
				RuleName:      "rule1",
				ReferenceType: "CARD",
			},
			expected: `uuid = 'abcd' AND name like '%rule1%' AND (JSON_CONTAINS(reference_type, '"CARD"') OR JSON_CONTAINS(reference_type, '"ANY"'))`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.query.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNew(t *testing.T) {
	// Store original generator and restore after tests
	originalGenerator := defaultUUIDGenerator
	defer func() {
		defaultUUIDGenerator = originalGenerator
	}()

	// Mock UUID for testing
	mockUUID := uuid.MustParse("f47ac10b-58cc-0372-8567-0e02b2c3d479")

	tests := []struct {
		name        string
		req         *CreateFraudRuleRequest
		uuidGen     UUIDGenerator
		expected    *FraudRules
		expectError bool
	}{
		{
			name: "successful creation",
			req: &CreateFraudRuleRequest{
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      5,
				Weight:        decimal.NewFromFloat(0.5),
				IsActive:      true,
				Provider:      sql.NullString{String: "INTERNAL", Valid: true},
				ReferenceType: "[\"ANY\"]",
			},
			uuidGen: func() (uuid.UUID, error) {
				return mockUUID, nil
			},
			expected: &FraudRules{
				UUID:          "f47ac10b-58cc-0372-8567-0e02b2c3d479",
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      5,
				Weight:        decimal.NewFromFloat(0.5),
				IsActive:      true,
				Provider:      sql.NullString{String: "INTERNAL", Valid: true},
				ReferenceType: "[\"ANY\"]",
			},
			expectError: false,
		},
		{
			name: "uuid generation failure",
			req: &CreateFraudRuleRequest{
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      5,
				Weight:        decimal.NewFromFloat(0.5),
				IsActive:      true,
				Provider:      sql.NullString{String: "INTERNAL", Valid: true},
				ReferenceType: "[\"ANY\"]",
			},
			uuidGen: func() (uuid.UUID, error) {
				return uuid.UUID{}, errors.New("uuid generation failed")
			},
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Set mock UUID generator
			defaultUUIDGenerator = tc.uuidGen

			// Execute the function under test
			result, err := New(tc.req)

			// Verify results
			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)

				// Check static fields
				assert.Equal(t, tc.expected.UUID, result.UUID)
				assert.Equal(t, tc.expected.RuleName, result.RuleName)
				assert.Equal(t, tc.expected.Condition, result.Condition)
				assert.Equal(t, tc.expected.Priority, result.Priority)
				assert.Equal(t, tc.expected.Weight, result.Weight)
				assert.Equal(t, tc.expected.IsActive, result.IsActive)
				assert.Equal(t, tc.expected.Provider, result.Provider)
				assert.Equal(t, tc.expected.ReferenceType, result.ReferenceType)

				// Check time fields are within a reasonable range
				assert.WithinDuration(t, time.Now().UTC(), result.CreatedAt, 2*time.Second)
				assert.WithinDuration(t, time.Now().UTC(), result.UpdatedAt, 2*time.Second)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	stringPtr := func(s string) *string { return &s }
	intPtr := func(i int) *int { return &i }
	boolPtr := func(b bool) *bool { return &b }
	decimalPtr := func(d float64) *decimal.Decimal {
		val := decimal.NewFromFloat(d)
		return &val
	}

	tests := []struct {
		name         string
		initialRule  *FraudRules
		updateReq    *UpdateFraudRuleRequest
		expectedRule *FraudRules
	}{
		{
			name: "update all fields",
			initialRule: &FraudRules{
				UUID:          "original-uuid",
				RuleName:      "Original Rule",
				Condition:     "amount > 500",
				Priority:      3,
				Weight:        decimal.NewFromFloat(0.5),
				IsActive:      false,
				Provider:      sql.NullString{String: "INTERNAL", Valid: true},
				ReferenceType: "[\"ANY\"]",
				CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			updateReq: &UpdateFraudRuleRequest{
				UUID:          "original-uuid",
				RuleName:      stringPtr("Updated Rule"),
				Condition:     stringPtr("amount > 1000"),
				Priority:      intPtr(5),
				Weight:        decimalPtr(1),
				IsActive:      boolPtr(true),
				Provider:      stringPtr("EXTERNAL"),
				ReferenceType: stringPtr("[\"VIRTUAL_ACCOUNT\"]"),
			},
			expectedRule: &FraudRules{
				UUID:          "original-uuid",
				RuleName:      "Updated Rule",
				Condition:     "amount > 1000",
				Priority:      5,
				Weight:        decimal.NewFromFloat(1),
				IsActive:      true,
				Provider:      sql.NullString{String: "EXTERNAL", Valid: true},
				ReferenceType: "[\"VIRTUAL_ACCOUNT\"]",
				CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "partial update with some fields unchanged",
			initialRule: &FraudRules{
				UUID:          "keep-uuid",
				RuleName:      "Original Rule",
				Condition:     "amount > 500",
				Priority:      3,
				Weight:        decimal.NewFromFloat(0.5),
				IsActive:      false,
				Provider:      sql.NullString{String: "EXTERNAL", Valid: true},
				ReferenceType: "[\"ANY\"]",
				CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			updateReq: &UpdateFraudRuleRequest{
				UUID:          "keep-uuid",
				Condition:     stringPtr("amount > 1000"),         // Changed
				Weight:        decimalPtr(1),                      // Changed
				IsActive:      boolPtr(true),                      // Changed
				ReferenceType: stringPtr("[\"VIRTUAL_ACCOUNT\"]"), // Changed
				// Others are nil (not changed)
			},
			expectedRule: &FraudRules{
				UUID:          "keep-uuid",
				RuleName:      "Original Rule", // Unchanged
				Condition:     "amount > 1000",
				Priority:      3, // Unchanged
				Weight:        decimal.NewFromFloat(1),
				IsActive:      true,
				Provider:      sql.NullString{String: "EXTERNAL", Valid: true}, // Unchanged
				ReferenceType: "[\"VIRTUAL_ACCOUNT\"]",
				CreatedAt:     time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			originalUpdatedAt := tc.initialRule.UpdatedAt

			time.Sleep(10 * time.Millisecond)

			tc.initialRule.Update(tc.updateReq)

			assert.Equal(t, tc.expectedRule.UUID, tc.initialRule.UUID)
			assert.Equal(t, tc.expectedRule.RuleName, tc.initialRule.RuleName)
			assert.Equal(t, tc.expectedRule.Condition, tc.initialRule.Condition)
			assert.Equal(t, tc.expectedRule.Priority, tc.initialRule.Priority)
			assert.True(t, tc.expectedRule.Weight.Equal(tc.initialRule.Weight))
			assert.Equal(t, tc.expectedRule.IsActive, tc.initialRule.IsActive)
			assert.Equal(t, tc.expectedRule.Provider, tc.initialRule.Provider)
			assert.Equal(t, tc.expectedRule.ReferenceType, tc.initialRule.ReferenceType)

			assert.Equal(t, tc.expectedRule.CreatedAt, tc.initialRule.CreatedAt)

			assert.True(t, tc.initialRule.UpdatedAt.After(originalUpdatedAt),
				"UpdatedAt should be more recent than original value")
			assert.WithinDuration(t, time.Now().UTC(), tc.initialRule.UpdatedAt, 2*time.Second)
		})
	}
}

func TestToResponse(t *testing.T) {
	testCases := []struct {
		name     string
		input    *FraudRules
		expected *FraudRulesResponse
	}{
		{
			name: "Complete fraud rule with valid fields",
			input: &FraudRules{
				UUID:          "01967b9c-6217-77e3-a754-531365a6f5f2",
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      sql.NullString{String: "test-provider", Valid: true},
				ReferenceType: "payment",
				CreatedAt:     time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
				DeletedAt:     sql.NullTime{Time: time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC), Valid: true},
			},
			expected: &FraudRulesResponse{
				UUID:          "01967b9c-6217-77e3-a754-531365a6f5f2",
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      "test-provider",
				ReferenceType: "payment",
				CreatedAt:     time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
				DeletedAt:     timePtr(time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC)),
			},
		},
		{
			name: "Fraud rule with null provider",
			input: &FraudRules{
				UUID:          "01967b9c-6217-77e3-a754-531365a6f5f2",
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      sql.NullString{String: "", Valid: false},
				ReferenceType: "payment",
				CreatedAt:     time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
				DeletedAt:     sql.NullTime{Valid: false},
			},
			expected: &FraudRulesResponse{
				UUID:          "01967b9c-6217-77e3-a754-531365a6f5f2",
				RuleName:      "Test Rule",
				Condition:     "amount > 1000",
				Priority:      1,
				Weight:        decimal.NewFromInt(50),
				IsActive:      true,
				Provider:      "",
				ReferenceType: "payment",
				CreatedAt:     time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt:     time.Date(2025, 4, 2, 0, 0, 0, 0, time.UTC),
				DeletedAt:     nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.input.ToResponse()

			assert.Equal(t, tc.expected.UUID, result.UUID)
			assert.Equal(t, tc.expected.RuleName, result.RuleName)
			assert.Equal(t, tc.expected.Condition, result.Condition)
			assert.Equal(t, tc.expected.Priority, result.Priority)
			assert.Equal(t, tc.expected.Weight.String(), result.Weight.String())
			assert.Equal(t, tc.expected.IsActive, result.IsActive)
			assert.Equal(t, tc.expected.Provider, result.Provider)
			assert.Equal(t, tc.expected.ReferenceType, result.ReferenceType)
			assert.Equal(t, tc.expected.CreatedAt, result.CreatedAt)
			assert.Equal(t, tc.expected.UpdatedAt, result.UpdatedAt)

			if tc.expected.DeletedAt == nil {
				assert.Nil(t, result.DeletedAt)
			} else {
				assert.NotNil(t, result.DeletedAt)
				assert.Equal(t, *tc.expected.DeletedAt, *result.DeletedAt)
			}
		})
	}
}

// Helper function to create a pointer to a time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
