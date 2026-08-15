package payoutManualProcessingAccount_test

import (
	"testing"

	. "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPayoutManualProcessingAccountQuery_BuildCondition(t *testing.T) {
	testMerchantID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")

	testCases := []struct {
		name         string
		query        *PayoutManualProcessingAccountQuery
		expectedSQL  string
		expectedArgs []any
	}{
		{
			name:         "Empty query",
			query:        &PayoutManualProcessingAccountQuery{},
			expectedSQL:  "",
			expectedArgs: []any{},
		},
		{
			name: "Query with MerchantID",
			query: &PayoutManualProcessingAccountQuery{
				MerchantID: testMerchantID,
			},
			expectedSQL:  "a.merchant_id = ?",
			expectedArgs: []any{testMerchantID.String()},
		},
		{
			name: "Query with BankCode",
			query: &PayoutManualProcessingAccountQuery{
				BankCode: "BCA",
			},
			expectedSQL:  "a.bank_code = ?",
			expectedArgs: []any{"BCA"},
		},
		{
			name: "Query with AccountNumber",
			query: &PayoutManualProcessingAccountQuery{
				AccountNumber: "1234567890",
			},
			expectedSQL:  "a.account_number = ?",
			expectedArgs: []any{"1234567890"},
		},
		{
			name: "Query with Status",
			query: &PayoutManualProcessingAccountQuery{
				Status: "ACTIVE",
			},
			expectedSQL:  "a.status = ?",
			expectedArgs: []any{"ACTIVE"},
		},
		{
			name: "Query with all fields",
			query: &PayoutManualProcessingAccountQuery{
				MerchantID:    testMerchantID,
				BankCode:      "BCA",
				AccountNumber: "1234567890",
				Status:        "ACTIVE",
			},
			expectedSQL:  "a.merchant_id = ? AND a.bank_code = ? AND a.account_number = ? AND a.status = ?",
			expectedArgs: []any{testMerchantID.String(), "BCA", "1234567890", "ACTIVE"},
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

func TestPayoutManualProcessingAccountQuery_BuildOrderBy(t *testing.T) {
	testCases := []struct {
		name     string
		query    *PayoutManualProcessingAccountQuery
		expected string
	}{
		{
			name:     "Empty sortBy and sort - use default",
			query:    &PayoutManualProcessingAccountQuery{},
			expected: "bank_code ASC",
		},
		{
			name: "Valid sortBy bankCode with ASC",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "bankCode",
				Sort:   "ASC",
			},
			expected: "bank_code ASC",
		},
		{
			name: "Valid sortBy accountNumber with DESC",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "accountNumber",
				Sort:   "DESC",
			},
			expected: "account_number DESC",
		},
		{
			name: "Valid sortBy status",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "status",
				Sort:   "ASC",
			},
			expected: "status ASC",
		},
		{
			name: "Valid sortBy uuid",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "uuid",
				Sort:   "DESC",
			},
			expected: "uuid DESC",
		},
		{
			name: "Invalid sortBy - use default",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "invalidField",
				Sort:   "ASC",
			},
			expected: "bank_code ASC",
		},
		{
			name: "Valid sortBy with invalid sort - use default ASC",
			query: &PayoutManualProcessingAccountQuery{
				SortBy: "bankCode",
				Sort:   "INVALID",
			},
			expected: "bank_code ASC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.query.BuildOrderBy()
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCreatePayoutManualProcessingAccountRequest_ToPayoutManualProcessingAccount(t *testing.T) {
	testCases := []struct {
		name string
		req  *CreatePayoutManualProcessingAccountRequest
	}{
		{
			name: "SUCCESS: Map request to account",
			req: &CreatePayoutManualProcessingAccountRequest{
				MerchantID:    "merchant-123",
				BankCode:      "BCA",
				AccountNumber: "1234567890",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.req.ToPayoutManualProcessingAccount()

			assert.NotNil(t, got)
			assert.NotEmpty(t, got.UUID)
			assert.Equal(t, tc.req.MerchantID, got.MerchantID)
			assert.Equal(t, tc.req.BankCode, got.BankCode)
			assert.Equal(t, tc.req.AccountNumber, got.AccountNumber)
			assert.NotEmpty(t, got.Status)
		})
	}
}

func TestPayoutManualProcessingAccount_Update(t *testing.T) {
	newStatus := "INACTIVE"

	testCases := []struct {
		name           string
		account        *PayoutManualProcessingAccount
		req            *UpdatePayoutManualProcessingAccountRequest
		expectedStatus string
	}{
		{
			name: "Update status",
			account: &PayoutManualProcessingAccount{
				UUID:   "uuid-123",
				Status: "ACTIVE",
			},
			req: &UpdatePayoutManualProcessingAccountRequest{
				UUID:   "uuid-123",
				Status: &newStatus,
			},
			expectedStatus: "INACTIVE",
		},
		{
			name: "Status nil - no change",
			account: &PayoutManualProcessingAccount{
				UUID:   "uuid-123",
				Status: "ACTIVE",
			},
			req: &UpdatePayoutManualProcessingAccountRequest{
				UUID:   "uuid-123",
				Status: nil,
			},
			expectedStatus: "ACTIVE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.account.Update(tc.req)
			assert.Equal(t, tc.expectedStatus, tc.account.Status)
		})
	}
}

func TestPayoutManualProcessingAccount_ToResponse(t *testing.T) {
	account := &PayoutManualProcessingAccount{
		UUID:          "uuid-123",
		MerchantID:    "merchant-123",
		BankCode:      "BCA",
		AccountNumber: "1234567890",
		Status:        "ACTIVE",
	}

	testCases := []struct {
		name    string
		account *PayoutManualProcessingAccount
	}{
		{
			name:    "SUCCESS: Map account to response",
			account: account,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.account.ToResponse()

			assert.NotNil(t, result)
			assert.Equal(t, tc.account.UUID, result.UUID)
			assert.Equal(t, tc.account.MerchantID, result.MerchantID)
			assert.Equal(t, tc.account.BankCode, result.BankCode)
			assert.Equal(t, tc.account.AccountNumber, result.AccountNumber)
			assert.Equal(t, tc.account.Status, result.Status)
		})
	}
}
