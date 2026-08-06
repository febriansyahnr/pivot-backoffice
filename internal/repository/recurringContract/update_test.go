package recurringContractRepo_test

import (
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	model "github.com/paper-indonesia/pivot-backoffice/internal/model/recurringContract"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/recurringContract"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateRecurringContractStatus(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(nil, db)

	uuid := "df5c965a-e5ff-4ea7-8875-2fe8082f0bc3"
	updatedBy := "559a94d4-09ba-4be0-a440-c8f32cd5a107"

	tests := []struct {
		name      string
		status    string
		updatedBy string
		setupMock func()
		wantError error
	}{
		{
			name:   "ERROR:Invalid status", // NOSONAR
			status: "INVALID_STATUS",
			setupMock: func() {
				// No database call should be made for invalid status
			},
			wantError: errors.New("invalid or unregistered status"),
		},
		{
			name:   "ERROR:No rows affected", // NOSONAR
			status: constant.RecurringContractStatusActive,
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusActive, mock.Anything, mock.Anything, uuid).Once().Return(false, nil)
			},
			wantError: constant.ErrNoRowsAffected,
		},
		{
			name:   "ERROR:Database error", // NOSONAR
			status: constant.RecurringContractStatusActive,
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusActive, mock.Anything, mock.Anything, uuid).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name:   "SUCCESS:Update to PENDING_INITIAL_AUTH without updatedBy", // NOSONAR
			status: constant.RecurringContractStatusPendInitialAuth,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusPendInitialAuth, mock.Anything, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?")
					require.Contains(t, query, "status IN ('CREATED', 'PENDING_INITIAL_AUTH')")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:      "SUCCESS:Update to PENDING_INITIAL_AUTH with updatedBy", // NOSONAR
			status:    constant.RecurringContractStatusPendInitialAuth,
			updatedBy: updatedBy,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusPendInitialAuth, mock.Anything, updatedBy, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?, updated_by = ?")
					require.Contains(t, query, "status IN ('CREATED', 'PENDING_INITIAL_AUTH')")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:   "SUCCESS:Update to ACTIVE without updatedBy", // NOSONAR
			status: constant.RecurringContractStatusActive,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusActive, mock.Anything, mock.Anything, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?, activated_at = ?")
					require.Contains(t, query, "status = 'PENDING_INITIAL_AUTH'")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:      "SUCCESS:Update to ACTIVE with updatedBy", // NOSONAR
			status:    constant.RecurringContractStatusActive,
			updatedBy: updatedBy,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusActive, mock.Anything, updatedBy, mock.Anything, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?, updated_by = ?, activated_at = ?")
					require.Contains(t, query, "status = 'PENDING_INITIAL_AUTH'")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:   "SUCCESS:Update to INACTIVE without updatedBy", // NOSONAR
			status: constant.RecurringContractStatusInactive,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusInactive, mock.Anything, mock.Anything, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?, deactivated_at = ?")
					require.Contains(t, query, "status IN ('CREATED', 'PENDING_INITIAL_AUTH', 'ACTIVE')")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name:      "SUCCESS:Update to INACTIVE with updatedBy", // NOSONAR
			status:    constant.RecurringContractStatusInactive,
			updatedBy: updatedBy,
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusInactive, mock.Anything, updatedBy, mock.Anything, uuid,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, updated_at = ?, updated_by = ?, deactivated_at = ?")
					require.Contains(t, query, "status IN ('CREATED', 'PENDING_INITIAL_AUTH', 'ACTIVE')")
				}).Return(true, nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateRecurringContractStatus(t.Context(), uuid, test.status, test.updatedBy)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
		})
	}
}

func TestUpdateRecurringContract(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(nil, db)

	recurringID := "df5c965a-e5ff-4ea7-8875-2fe8082f0bc3"
	transactionID := "TRX-123456"    // NOSONAR
	paymentTokenID := "TOKEN-123456" // NOSONAR
	paymentMethodID := "PM-123456"   // NOSONAR
	updatedBy := "559a94d4-09ba-4be0-a440-c8f32cd5a107"
	updatedAt := time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC)
	activatedAt := time.Date(2025, 1, 13, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		payload   model.UpdateRecurringContractRequest
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Empty parameters", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID: recurringID,
			},
			setupMock: func() {
				// No database call should be made for empty parameters
			},
			wantError: errors.New("update parameters must not be empty"),
		},
		{
			name: "ERROR:Database error", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:   recurringID,
				TransactionID: transactionID,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID, recurringID).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:No rows affected", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:   recurringID,
				TransactionID: transactionID,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID, recurringID).Once().Return(false, nil)
			},
			wantError: constant.ErrNoRowsAffected,
		},
		{
			name: "SUCCESS:Update BillingCycleCount only", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:       recurringID,
				BillingCycleCount: 5,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, uint16(5), recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "billing = JSON_SET(billing, '$.count', ?) WHERE uuid = ?;")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update Status and ActivatedAt", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID: recurringID,
				Status:      constant.RecurringContractStatusActive,
				ActivatedAt: activatedAt,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, constant.RecurringContractStatusActive, activatedAt, recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "status = ?, activated_at = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update TransactionID only", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:   recurringID,
				TransactionID: transactionID,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, transactionID, recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "auth_transaction_id = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update PaymentTokenID only", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:    recurringID,
				PaymentTokenID: paymentTokenID,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, paymentTokenID, recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "payment_token_id = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update PaymentMethodID only", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:     recurringID,
				PaymentMethodID: paymentMethodID,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, paymentMethodID, recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "payment_method_id = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update UpdatedAt and UpdatedBy only", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID: recurringID,
				UpdatedAt:   updatedAt,
				UpdatedBy:   updatedBy,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, updatedAt, updatedBy, recurringID).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "updated_at = ?, updated_by = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update all fields", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:       recurringID,
				BillingCycleCount: 5,
				Status:            constant.RecurringContractStatusActive,
				ActivatedAt:       activatedAt,
				TransactionID:     transactionID,
				PaymentTokenID:    paymentTokenID,
				PaymentMethodID:   paymentMethodID,
				UpdatedAt:         updatedAt,
				UpdatedBy:         updatedBy,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything,
					uint16(5),
					constant.RecurringContractStatusActive,
					activatedAt,
					transactionID,
					paymentTokenID,
					paymentMethodID,
					updatedAt,
					updatedBy,
					recurringID,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "billing = JSON_SET(billing, '$.count', ?), status = ?, activated_at = ?, auth_transaction_id = ?, payment_token_id = ?, payment_method_id = ?, updated_at = ?, updated_by = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Update multiple fields", // NOSONAR
			payload: model.UpdateRecurringContractRequest{
				RecurringID:    recurringID,
				TransactionID:  transactionID,
				PaymentTokenID: paymentTokenID,
				UpdatedBy:      updatedBy,
			},
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything,
					transactionID,
					paymentTokenID,
					updatedBy,
					recurringID,
				).Once().Run(func(args mock.Arguments) {
					query := args.Get(1).(string)
					require.Contains(t, query, "auth_transaction_id = ?, payment_token_id = ?, updated_by = ? WHERE uuid = ?")
				}).Return(true, nil)
			},
			wantError: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdateRecurringContract(t.Context(), test.payload)
			assert.Equal(t, test.wantError, err)

			db.AssertExpectations(t)
		})
	}
}
