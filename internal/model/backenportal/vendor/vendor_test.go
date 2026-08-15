package vendor

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestCreateVendorRequest_ToVendor(t *testing.T) {
	testCases := []struct {
		name  string
		input *CreateVendorRequest
	}{
		{
			name: "SUCCESS: Create new vendor",
			input: &CreateVendorRequest{
				MerchantID:          "merchant-123",
				Name:                "Test Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
			},
		},
		{
			name: "SUCCESS: Create new vendor with documents",
			input: &CreateVendorRequest{
				MerchantID:          "merchant-123",
				Name:                "Test Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
				Documents:           types.JSONText(`{"doc": "value"}`),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.ToVendor()

			assert.NotNil(t, got)
			assert.NotEmpty(t, got.UUID)
			assert.Equal(t, tc.input.MerchantID, got.MerchantID)
			assert.Equal(t, tc.input.Name, got.Name)
			assert.Equal(t, tc.input.BeneficialOwner, got.BeneficialOwner)
			assert.Equal(t, tc.input.BusinessCategory, got.BusinessCategory)
			assert.Equal(t, tc.input.AvgMonthlyTpvAmount.String(), got.AvgMonthlyTpvAmount.String())
			assert.Equal(t, tc.input.BankName, got.BankName)
			assert.Equal(t, tc.input.BankCode, got.BankCode)
			assert.Equal(t, tc.input.AccountNumber, got.AccountNumber)
			assert.Equal(t, tc.input.AccountName, got.AccountName)
			assert.Equal(t, "ACTIVE", got.Status)
			assert.False(t, got.CreatedAt.IsZero())
			assert.False(t, got.UpdatedAt.IsZero())
		})
	}
}

func TestVendor_Update(t *testing.T) {
	vendor := &Vendor{
		UUID:                "test-uuid",
		MerchantID:          "merchant-123",
		Name:                "Original Name",
		BeneficialOwner:     "Original Owner",
		BusinessCategory:    "Original Category",
		AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
		BankName:            "Original Bank",
		BankCode:            "ORI",
		AccountNumber:       "1111111111",
		AccountName:         "Original Account",
		Status:              "ACTIVE",
		CreatedAt:           time.Now().Add(-time.Hour),
		UpdatedAt:           time.Now().Add(-time.Hour),
	}

	newName := "Updated Name"
	newOwner := "Updated Owner"
	newCategory := "Updated Category"
	newAmount := decimal.NewFromInt(2000000)
	newBankName := "Updated Bank"
	newBankCode := "UPD"
	newAccountNumber := "2222222222"
	newAccountName := "Updated Account"
	newDocuments := types.JSONText(`{"updated": true}`)

	testCases := []struct {
		name     string
		request  *UpdateVendorRequest
		expected func(v *Vendor) bool
	}{
		{
			name: "Update name only",
			request: &UpdateVendorRequest{
				UUID: "test-uuid",
				Name: &newName,
			},
			expected: func(v *Vendor) bool {
				return v.Name == newName
			},
		},
		{
			name: "Update beneficial owner only",
			request: &UpdateVendorRequest{
				UUID:            "test-uuid",
				BeneficialOwner: &newOwner,
			},
			expected: func(v *Vendor) bool {
				return v.BeneficialOwner == newOwner
			},
		},
		{
			name: "Update business category only",
			request: &UpdateVendorRequest{
				UUID:             "test-uuid",
				BusinessCategory: &newCategory,
			},
			expected: func(v *Vendor) bool {
				return v.BusinessCategory == newCategory
			},
		},
		{
			name: "Update avg monthly tpv amount only",
			request: &UpdateVendorRequest{
				UUID:                "test-uuid",
				AvgMonthlyTpvAmount: &newAmount,
			},
			expected: func(v *Vendor) bool {
				return v.AvgMonthlyTpvAmount.Equal(newAmount)
			},
		},
		{
			name: "Update bank name only",
			request: &UpdateVendorRequest{
				UUID:     "test-uuid",
				BankName: &newBankName,
			},
			expected: func(v *Vendor) bool {
				return v.BankName == newBankName
			},
		},
		{
			name: "Update bank code only",
			request: &UpdateVendorRequest{
				UUID:     "test-uuid",
				BankCode: &newBankCode,
			},
			expected: func(v *Vendor) bool {
				return v.BankCode == newBankCode
			},
		},
		{
			name: "Update account number only",
			request: &UpdateVendorRequest{
				UUID:          "test-uuid",
				AccountNumber: &newAccountNumber,
			},
			expected: func(v *Vendor) bool {
				return v.AccountNumber == newAccountNumber
			},
		},
		{
			name: "Update account name only",
			request: &UpdateVendorRequest{
				UUID:        "test-uuid",
				AccountName: &newAccountName,
			},
			expected: func(v *Vendor) bool {
				return v.AccountName == newAccountName
			},
		},
		{
			name: "Update documents only",
			request: &UpdateVendorRequest{
				UUID:      "test-uuid",
				Documents: &newDocuments,
			},
			expected: func(v *Vendor) bool {
				return string(v.Documents) == string(newDocuments)
			},
		},
		{
			name: "Update all fields",
			request: &UpdateVendorRequest{
				UUID:                "test-uuid",
				Name:                &newName,
				BeneficialOwner:     &newOwner,
				BusinessCategory:    &newCategory,
				AvgMonthlyTpvAmount: &newAmount,
				BankName:            &newBankName,
				BankCode:            &newBankCode,
				AccountNumber:       &newAccountNumber,
				AccountName:         &newAccountName,
				Documents:           &newDocuments,
			},
			expected: func(v *Vendor) bool {
				return v.Name == newName &&
					v.BeneficialOwner == newOwner &&
					v.BusinessCategory == newCategory &&
					v.AvgMonthlyTpvAmount.Equal(newAmount) &&
					v.BankName == newBankName &&
					v.BankCode == newBankCode &&
					v.AccountNumber == newAccountNumber &&
					v.AccountName == newAccountName &&
					string(v.Documents) == string(newDocuments)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a fresh copy for each test
			v := &Vendor{
				UUID:                vendor.UUID,
				MerchantID:          vendor.MerchantID,
				Name:                vendor.Name,
				BeneficialOwner:     vendor.BeneficialOwner,
				BusinessCategory:    vendor.BusinessCategory,
				AvgMonthlyTpvAmount: vendor.AvgMonthlyTpvAmount,
				BankName:            vendor.BankName,
				BankCode:            vendor.BankCode,
				AccountNumber:       vendor.AccountNumber,
				AccountName:         vendor.AccountName,
				Status:              vendor.Status,
				CreatedAt:           vendor.CreatedAt,
				UpdatedAt:           vendor.UpdatedAt,
			}

			oldUpdatedAt := v.UpdatedAt

			v.Update(tc.request)

			assert.True(t, tc.expected(v))
			assert.True(t, v.UpdatedAt.After(oldUpdatedAt) || v.UpdatedAt.Equal(oldUpdatedAt))
		})
	}
}

func TestVendor_ToResponse(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name   string
		vendor *Vendor
		check  func(r *VendorResponse)
	}{
		{
			name: "Vendor without deleted_at",
			vendor: &Vendor{
				UUID:                "test-uuid",
				MerchantID:          "merchant-123",
				Name:                "Test Vendor",
				BeneficialOwner:     "John Doe",
				BusinessCategory:    "E-Commerce",
				AvgMonthlyTpvAmount: decimal.NewFromInt(1000000),
				BankName:            "Bank ABC",
				BankCode:            "ABC",
				AccountNumber:       "1234567890",
				AccountName:         "Test Account",
				Documents:           types.JSONText(`{"doc": "value"}`),
				Status:              "ACTIVE",
				CreatedAt:           now,
				UpdatedAt:           now,
				DeletedAt:           sql.NullTime{Valid: false},
			},
			check: func(r *VendorResponse) {
				assert.Equal(t, "test-uuid", r.UUID)
				assert.Equal(t, "merchant-123", r.MerchantID)
				assert.Equal(t, "Test Vendor", r.Name)
				assert.Equal(t, "John Doe", r.BeneficialOwner)
				assert.Equal(t, "E-Commerce", r.BusinessCategory)
				assert.Equal(t, decimal.NewFromInt(1000000).String(), r.AvgMonthlyTpvAmount.String())
				assert.Equal(t, "Bank ABC", r.BankName)
				assert.Equal(t, "ABC", r.BankCode)
				assert.Equal(t, "1234567890", r.AccountNumber)
				assert.Equal(t, "Test Account", r.AccountName)
				assert.Equal(t, "ACTIVE", r.Status)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.vendor.ToResponse()
			tc.check(result)
		})
	}
}

func TestVendorQuery_BuildCondition(t *testing.T) {
	testUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	testMerchantID := uuid.MustParse("660e8400-e29b-41d4-a716-446655440001")
	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)

	testCases := []struct {
		name         string
		query        *VendorQuery
		expectedSQL  string
		expectedArgs []any
	}{
		{
			name:         "Empty query",
			query:        &VendorQuery{},
			expectedSQL:  "deleted_at IS NULL",
			expectedArgs: []any{},
		},
		{
			name: "Query with UUID",
			query: &VendorQuery{
				UUID: testUUID,
			},
			expectedSQL:  "uuid = ? AND deleted_at IS NULL",
			expectedArgs: []any{testUUID.String()},
		},
		{
			name: "Query with MerchantID",
			query: &VendorQuery{
				MerchantID: testMerchantID,
			},
			expectedSQL:  "merchant_id = ? AND deleted_at IS NULL",
			expectedArgs: []any{testMerchantID.String()},
		},
		{
			name: "Query with Name",
			query: &VendorQuery{
				Name: "Test",
			},
			expectedSQL:  "name LIKE ? AND deleted_at IS NULL",
			expectedArgs: []any{"%Test%"},
		},
		{
			name: "Query with Status",
			query: &VendorQuery{
				Status: "ACTIVE",
			},
			expectedSQL:  "status = ? AND deleted_at IS NULL",
			expectedArgs: []any{"ACTIVE"},
		},
		{
			name: "Query with StartDate only",
			query: &VendorQuery{
				StartDate: &startDate,
			},
			expectedSQL:  "created_at >= ? AND deleted_at IS NULL",
			expectedArgs: []any{startDate},
		},
		{
			name: "Query with EndDate only",
			query: &VendorQuery{
				EndDate: &endDate,
			},
			expectedSQL:  "created_at <= ? AND deleted_at IS NULL",
			expectedArgs: []any{endDate},
		},
		{
			name: "Query with StartDate and EndDate",
			query: &VendorQuery{
				StartDate: &startDate,
				EndDate:   &endDate,
			},
			expectedSQL:  "created_at >= ? AND created_at <= ? AND deleted_at IS NULL",
			expectedArgs: []any{startDate, endDate},
		},
		{
			name: "Query with all fields",
			query: &VendorQuery{
				UUID:       testUUID,
				MerchantID: testMerchantID,
				Name:       "Test",
				Status:     "ACTIVE",
			},
			expectedSQL:  "uuid = ? AND merchant_id = ? AND name LIKE ? AND status = ? AND deleted_at IS NULL",
			expectedArgs: []any{testUUID.String(), testMerchantID.String(), "%Test%", "ACTIVE"},
		},
		{
			name: "Query with all fields including dates",
			query: &VendorQuery{
				UUID:       testUUID,
				MerchantID: testMerchantID,
				Name:       "Test",
				Status:     "ACTIVE",
				StartDate:  &startDate,
				EndDate:    &endDate,
			},
			expectedSQL:  "uuid = ? AND merchant_id = ? AND name LIKE ? AND status = ? AND created_at >= ? AND created_at <= ? AND deleted_at IS NULL",
			expectedArgs: []any{testUUID.String(), testMerchantID.String(), "%Test%", "ACTIVE", startDate, endDate},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sql, args := tc.query.BuildCondition()
			assert.Equal(t, tc.expectedSQL, sql)
			assert.Equal(t, tc.expectedArgs, args)
		})
	}
}

func TestVendorQuery_BuildOrderBy(t *testing.T) {
	testCases := []struct {
		name     string
		query    *VendorQuery
		expected string
	}{
		{
			name:     "Empty sortBy and sort - use default",
			query:    &VendorQuery{},
			expected: "created_at DESC",
		},
		{
			name: "Valid sortBy createdAt with ASC",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "ASC",
			},
			expected: "created_at ASC",
		},
		{
			name: "Valid sortBy createdAt with DESC",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "DESC",
			},
			expected: "created_at DESC",
		},
		{
			name: "Valid sortBy createdAt with lowercase asc",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "asc",
			},
			expected: "created_at ASC",
		},
		{
			name: "Valid sortBy createdAt with lowercase desc",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "desc",
			},
			expected: "created_at DESC",
		},
		{
			name: "Invalid sortBy - use default",
			query: &VendorQuery{
				SortBy: "invalidField",
				Sort:   "ASC",
			},
			expected: "created_at DESC",
		},
		{
			name: "Valid sortBy with invalid sort - use default DESC",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "INVALID",
			},
			expected: "created_at DESC",
		},
		{
			name: "Empty sortBy with valid sort - use default",
			query: &VendorQuery{
				SortBy: "",
				Sort:   "ASC",
			},
			expected: "created_at DESC",
		},
		{
			name: "Valid sortBy with empty sort - use default DESC",
			query: &VendorQuery{
				SortBy: "createdAt",
				Sort:   "",
			},
			expected: "created_at DESC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.query.BuildOrderBy()
			assert.Equal(t, tc.expected, result)
		})
	}
}
