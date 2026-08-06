package vccSettlement

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	vccSettlement "github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBulkInsert(t *testing.T) {
	now := time.Now()
	sourceAmount, _ := json.Marshal(map[string]any{"value": "100.00", "currency": "USD"})
	billingAmount, _ := json.Marshal(map[string]any{"value": "1500000.00", "currency": "IDR"})

	validData := []*vccSettlement.VccSettlement{
		{
			UUID:                    uuid.New().String(),
			RcnId:                   uuid.New().String(),
			AcquirerReferenceNumber: "ARN123456789",
			Status:                  "SETTLED",
			ReferenceNo:             "REF123",
			AuthorizationNo:         "AUTH456",
			PostingDate:             now,
			BillingCycle:            202501,
			SourceAmount:            sourceAmount,
			BillingAmount:           billingAmount,
			TransactionDate:         now,
			SettlementDate:          now.Add(24 * time.Hour),
			MerchantName:            "Test Merchant",
			MerchantCountry:         "US",
			MerchantCategory:        "5411",
			CreatedAt:               now,
			UpdatedAt:               now,
			DeletedAt:               sql.NullTime{Valid: false},
		},
		{
			UUID:                    uuid.New().String(),
			RcnId:                   uuid.New().String(),
			AcquirerReferenceNumber: "ARN987654321",
			Status:                  "PENDING",
			ReferenceNo:             "REF789",
			AuthorizationNo:         "AUTH999",
			PostingDate:             now,
			BillingCycle:            202501,
			SourceAmount:            sourceAmount,
			BillingAmount:           billingAmount,
			TransactionDate:         now,
			SettlementDate:          now.Add(48 * time.Hour),
			MerchantName:            "Another Merchant",
			MerchantCountry:         "ID",
			MerchantCategory:        "5812",
			CreatedAt:               now,
			UpdatedAt:               now,
			DeletedAt:               sql.NullTime{Valid: false},
		},
	}

	testCases := []struct {
		name      string
		input     []*vccSettlement.VccSettlement
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		assertFn  func(t *testing.T, err error)
	}{
		{
			name:  "Success - Insert single record",
			input: validData[:1],
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Expect 18 arguments for 1 record (after ctx and query)
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "Success - Insert multiple records",
			input: validData,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Expect 36 arguments for 2 records (18 cols * 2 rows)
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					// First record (18 fields)
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
					// Second record (18 fields)
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:  "Failure - Database error during insert",
			input: validData[:1],
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(false, errors.New("database connection error")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "database connection error")
			},
		},
		{
			name:  "Failure - No rows affected",
			input: validData[:1],
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(false, nil).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "no rows affected")
			},
		},
		{
			name:  "Success - Empty input slice",
			input: []*vccSettlement.VccSettlement{},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// For empty input, ExecContext should still be called but with no data args
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name: "Success - Record with deleted_at set",
			input: []*vccSettlement.VccSettlement{
				{
					UUID:                    uuid.New().String(),
					RcnId:                   uuid.New().String(),
					AcquirerReferenceNumber: "ARN111",
					Status:                  "DELETED",
					ReferenceNo:             "REF111",
					AuthorizationNo:         "AUTH111",
					PostingDate:             now,
					BillingCycle:            202501,
					SourceAmount:            sourceAmount,
					BillingAmount:           billingAmount,
					TransactionDate:         now,
					SettlementDate:          now,
					MerchantName:            "Deleted Merchant",
					MerchantCountry:         "US",
					MerchantCategory:        "5411",
					CreatedAt:               now,
					UpdatedAt:               now,
					DeletedAt:               sql.NullTime{Time: now, Valid: true},
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything,
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)

			// Execute
			err := repo.BulkInsert(context.Background(), tc.input)

			// Assert
			if tc.wantErr {
				require.Error(t, err)
				if tc.assertFn != nil {
					tc.assertFn(t, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBulkInsert_QueryStructure(t *testing.T) {
	// This test verifies the query structure is correct
	now := time.Now()
	sourceAmount, _ := json.Marshal(map[string]interface{}{"value": "100.00", "currency": "USD"})
	billingAmount, _ := json.Marshal(map[string]interface{}{"value": "1500000.00", "currency": "IDR"})

	input := []*vccSettlement.VccSettlement{
		{
			UUID:                    "test-uuid",
			RcnId:                   "test-rcn-id",
			AcquirerReferenceNumber: "ARN123",
			Status:                  "SETTLED",
			ReferenceNo:             "REF123",
			AuthorizationNo:         "AUTH123",
			PostingDate:             now,
			BillingCycle:            202501,
			SourceAmount:            sourceAmount,
			BillingAmount:           billingAmount,
			TransactionDate:         now,
			SettlementDate:          now,
			MerchantName:            "Test",
			MerchantCountry:         "US",
			MerchantCategory:        "5411",
			CreatedAt:               now,
			UpdatedAt:               now,
			DeletedAt:               sql.NullTime{Valid: false},
		},
	}

	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	// Capture the query
	var capturedQuery string
	mockMysql.On(
		"ExecContext",
		mock.Anything,
		mock.MatchedBy(func(query string) bool {
			capturedQuery = query
			return true
		}),
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything,
	).Return(true, nil).Once()

	repo := New(mockMysql, mockLogger)
	err := repo.BulkInsert(context.Background(), input)

	require.NoError(t, err)
	assert.Contains(t, capturedQuery, "INSERT INTO vcc_settlements")
	assert.Contains(t, capturedQuery, "uuid")
	assert.Contains(t, capturedQuery, "rcn_id")
	assert.Contains(t, capturedQuery, "acquirer_reference_number")
	assert.Contains(t, capturedQuery, "status")
	assert.Contains(t, capturedQuery, "VALUES")
	assert.Contains(t, capturedQuery, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
}
