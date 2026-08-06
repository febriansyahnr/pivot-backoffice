package vccSettlement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	now := time.Now()
	validRcnId := uuid.New().String()

	testCases := []struct {
		name        string
		rcnId       string
		postingDate time.Time
		mockSetup   func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr     bool
		assertFn    func(t *testing.T, err error)
	}{
		{
			name:        "Success - Delete with valid rcnId and postingDate",
			rcnId:       validRcnId,
			postingDate: now,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Expect 3 arguments: time.Now(), postingDate, rcnId
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"), // deleted_at timestamp
					mock.AnythingOfType("time.Time"), // postingDate
					mock.AnythingOfType("string"),    // rcnId
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Success - Delete with different rcnId",
			rcnId:       "different-rcn-id",
			postingDate: now.Add(-24 * time.Hour),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Failure - Database error during delete",
			rcnId:       validRcnId,
			postingDate: now,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("database connection error")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "database connection error")
			},
		},
		{
			name:        "Failure - Constraint violation error",
			rcnId:       validRcnId,
			postingDate: now,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("constraint violation")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "constraint violation")
			},
		},
		{
			name:        "Success - Delete with empty rcnId (no rows affected but no error)",
			rcnId:       "",
			postingDate: now,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				// Note: The implementation doesn't check rows affected,
				// so even with empty rcnId, it returns success if DB exec succeeds
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(false, nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Success - Delete with zero time",
			rcnId:       validRcnId,
			postingDate: time.Time{},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Success - Delete with future postingDate",
			rcnId:       validRcnId,
			postingDate: now.Add(365 * 24 * time.Hour),
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(true, nil).Once()
			},
			wantErr: false,
		},
		{
			name:        "Failure - Context deadline exceeded",
			rcnId:       validRcnId,
			postingDate: now,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("time.Time"),
					mock.AnythingOfType("string"),
				).Return(false, errors.New("context deadline exceeded")).Once()
			},
			wantErr: true,
			assertFn: func(t *testing.T, err error) {
				assert.Contains(t, err.Error(), "context deadline exceeded")
			},
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
			err := repo.Delete(context.Background(), tc.rcnId, tc.postingDate)

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

func TestDelete_QueryStructure(t *testing.T) {
	// This test verifies the query structure is correct
	now := time.Now()
	rcnId := uuid.New().String()

	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	// Capture the query and arguments
	var capturedQuery string
	var capturedDeletedAt time.Time
	var capturedPostingDate time.Time
	var capturedRcnId string

	mockMysql.On(
		"ExecContext",
		mock.Anything,
		mock.MatchedBy(func(query string) bool {
			capturedQuery = query
			return true
		}),
		mock.MatchedBy(func(deletedAt time.Time) bool {
			capturedDeletedAt = deletedAt
			return true
		}),
		mock.MatchedBy(func(postingDate time.Time) bool {
			capturedPostingDate = postingDate
			return true
		}),
		mock.MatchedBy(func(id string) bool {
			capturedRcnId = id
			return true
		}),
	).Return(true, nil).Once()

	repo := New(mockMysql, mockLogger)
	err := repo.Delete(context.Background(), rcnId, now)

	require.NoError(t, err)

	// Verify query structure
	assert.Contains(t, capturedQuery, "UPDATE vcc_settlements")
	assert.Contains(t, capturedQuery, "SET deleted_at = ?")
	assert.Contains(t, capturedQuery, "WHERE posting_date = ? AND rcn_id = ?")

	// Verify deleted_at is set to current time (approximately)
	assert.WithinDuration(t, time.Now(), capturedDeletedAt, 2*time.Second, "deleted_at should be set to current time")

	// Verify postingDate matches input
	assert.Equal(t, now, capturedPostingDate, "postingDate should match input")

	// Verify rcnId matches input
	assert.Equal(t, rcnId, capturedRcnId, "rcnId should match input")
}

func TestDelete_ArgumentOrder(t *testing.T) {
	// This test ensures arguments are passed in the correct order
	now := time.Date(2026, 2, 4, 10, 0, 0, 0, time.UTC)
	rcnId := "test-rcn-123"

	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	mockMysql := mysqlMocks.NewIMySqlExt(t)

	// Capture arguments individually
	var capturedDeletedAt time.Time
	var capturedPostingDate time.Time
	var capturedRcnId string

	mockMysql.On(
		"ExecContext",
		mock.Anything,
		mock.Anything,
		mock.MatchedBy(func(deletedAt time.Time) bool {
			capturedDeletedAt = deletedAt
			return true
		}),
		mock.MatchedBy(func(postingDate time.Time) bool {
			capturedPostingDate = postingDate
			return true
		}),
		mock.MatchedBy(func(id string) bool {
			capturedRcnId = id
			return true
		}),
	).Return(true, nil).Once()

	repo := New(mockMysql, mockLogger)
	err := repo.Delete(context.Background(), rcnId, now)

	require.NoError(t, err)

	// Verify deleted_at is set to current time (approximately)
	assert.WithinDuration(t, time.Now(), capturedDeletedAt, 2*time.Second, "deleted_at should be set to current time")

	// Verify postingDate matches input
	assert.Equal(t, now, capturedPostingDate, "postingDate should match input")

	// Verify rcnId matches input
	assert.Equal(t, rcnId, capturedRcnId, "rcnId should match input")
}