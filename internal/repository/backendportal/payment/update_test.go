package paymentRepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentRepositoryUpdatePayment(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					mock.Anything,
					mock.Anything,
					constant.StringMockType(),
					constant.StringMockType(),
					constant.TimeMockType(),
					constant.StringMockType(),
				).Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "payments")
			err := repo.UpdatePayment(
				ctx,
				uuid.NewString(),
				decimal.NewFromInt(1000000), decimal.NewFromInt(1000000), "metadata", uuid.NewString(), util.TimeNow)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepositoryUpdatePaymentStatus(t *testing.T) {
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(false, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "payments")
			err := repo.UpdatePaymentStatus(ctx, uuid.NewString(), uuid.NewString(), paymentConstant.PAYMENT_STATUS_SUCCESS, util.TimeNow)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepositoryPaymentItemsFromPaymentResponseItem(t *testing.T) {
	paymentRespItems := []paymentModel.PaymentResponseItem{
		{
			ItemID: uuid.New().String(),
			Name:   "item 1",
		},
		{
			ItemID: "",
			Name:   "item 2",
		},
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update payment items from payment response item",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"Commit",
					mock.Anything,
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Error when begin tx transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(nil, errors.New("begin tx error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error when delete payment items by payment id",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, fmt.Errorf("error when deleting payment items by payment id"))
				mysqlMock.On(
					"Rollback",
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "FAILED: Error when insert payment items",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, fmt.Errorf("error when inserting payment items"))
				mysqlMock.On(
					"Rollback",
					mock.Anything,
				).Return(nil)
			},
			wantErr: true,
		},
		{
			name: "FAILED: failed when commit transaction",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"Commit",
					mock.Anything,
				).Return(fmt.Errorf("error when commit transaction"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error when delete payment items with rollback error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, fmt.Errorf("error when deleting payment items by payment id"))
				mysqlMock.On(
					"Rollback",
					mock.Anything,
				).Return(fmt.Errorf("rollback failed"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Error when insert payment items with rollback error",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				tx := &sqlx.Tx{}
				ctx := context.WithValue(context.Background(), pdkConst.CtxSqlTx, tx)
				mysqlMock.On(
					"BeginTxx",
					mock.Anything,
				).Return(ctx, nil)
				mysqlMock.On(
					"ExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(true, nil)
				mysqlMock.On(
					"NamedExecContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(false, fmt.Errorf("error when inserting payment items"))
				mysqlMock.On(
					"Rollback",
					mock.Anything,
				).Return(fmt.Errorf("rollback failed"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "payment_items")
			err := repo.UpdatePaymentItemsFromPaymentResponseItem(ctx, uuid.NewString(), paymentRespItems)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestPaymentRepositoryUpdatePaymentData(t *testing.T) {
	now := time.Now()
	expiredAt := time.Now().Add(constant.CreditCardPaymentExpired)
	referenceId := "some-reference-id"
	paymentDB := &paymentModel.Payment{
		UUID:        uuid.NewString(),
		ReferenceID: &referenceId,
		Amount:      decimal.NewFromFloat(10000),
		Currency:    "IDR",
		PaymentURL:  "https://creditcard-webview-stg.harsya.com/payment/creditcard/pay/hOIuKiu-6NxhiFnJPMDWIke9qq0YsbpERh4Atnn-AEY=",
		Status:      "WAITING_FOR_PAYMENT",
		ExpiredAt:   &expiredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: No rows affected",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, constant.ErrNoRowsAffected)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Failure update payment status",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContextReturnLastId",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, errors.New("some error"))
			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "payments")
			err := repo.UpdatePaymentData(ctx, paymentDB.ToDTO())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestUpdatePaymentMetadataById(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	paymentId := "6fc37d33-ab12-47c7-8e38-8d7c15b95368"

	tests := []struct {
		name      string
		request   paymentModel.UpdatePaymentMetadataRequest
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Some error",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FeeDetail:   map[string]string{},
				FeeOnBehalf: map[string]string{},
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(false, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS - FeeDetail and FeeOnBehalf",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FeeDetail:   map[string]string{"test": "value"},
				FeeOnBehalf: map[string]string{"behalf": "data"},
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - FingerprintID only",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FingerprintID: "test-fingerprint-123",
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - SummaryTransaction only",
			request: paymentModel.UpdatePaymentMetadataRequest{
				SummaryTransaction: map[string]interface{}{
					"total":    1000,
					"currency": "IDR",
					"tax":      100,
					"discount": 50,
				},
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - All fields populated",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FeeDetail: map[string]string{
					"type":   "percentage",
					"amount": "2.5",
				},
				FeeOnBehalf: map[string]string{
					"payer": "merchant",
					"type":  "fixed",
				},
				SummaryTransaction: map[string]interface{}{
					"subtotal": 10000,
					"tax":      1000,
					"total":    11000,
				},
				FingerprintID: "fingerprint-abc123def456",
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name:    "SUCCESS - Empty metadata (no fields to update)",
			request: paymentModel.UpdatePaymentMetadataRequest{
				// All fields are nil
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "ERROR - FingerprintID update fails",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FingerprintID: "test-fingerprint-456",
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(false, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "ERROR - SummaryTransaction update fails",
			request: paymentModel.UpdatePaymentMetadataRequest{
				SummaryTransaction: map[string]interface{}{
					"amount": 5000,
					"fee":    250,
				},
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(false, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS - Mixed field combination",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FeeDetail:     map[string]string{"rate": "3.0"},
				FingerprintID: "mixed-test-fingerprint",
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - IsSnap true",
			request: paymentModel.UpdatePaymentMetadataRequest{
				IsSnap: func() *bool { b := true; return &b }(),
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - IsSnap false",
			request: paymentModel.UpdatePaymentMetadataRequest{
				IsSnap: func() *bool { b := false; return &b }(),
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
		{
			name: "SUCCESS - All fields including IsSnap",
			request: paymentModel.UpdatePaymentMetadataRequest{
				FeeDetail: map[string]string{
					"type":   "percentage",
					"amount": "2.5",
				},
				FeeOnBehalf: map[string]string{
					"payer": "merchant",
					"type":  "fixed",
				},
				SummaryTransaction: map[string]interface{}{
					"subtotal": 10000,
					"tax":      1000,
					"total":    11000,
				},
				FingerprintID: "fingerprint-abc123def456",
				IsSnap:        func() *bool { b := true; return &b }(),
			},
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, paymentId,
				).Once().Return(true, nil)
			},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := repo.UpdatePaymentMetadataById(context.Background(), paymentId, test.request)
			assert.Equal(t, test.wantErr, err)
		})
	}
}

func TestPaymentRepositoryUpdatePaymentForInvestigation(t *testing.T) {
	testCase := []struct {
		name      string
		request   paymentModel.UpdatePaymentForInvestigationRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update payment for investigation",
			request: paymentModel.UpdatePaymentForInvestigationRequest{
				PaymentID:  uuid.NewString(),
				MerchantID: uuid.NewString(),
				ReasonType: "INVESTIGATION_IN_PROCESS",
				InvestigationMetadata: paymentModel.InvestigationPoPMetadata{
					Bucket:        "test-bucket",
					Path:          "investigations/merchant-123/payment-456.png",
					MerchantNotes: "Customer showed screenshot",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with empty metadata",
			request: paymentModel.UpdatePaymentForInvestigationRequest{
				PaymentID:             uuid.NewString(),
				MerchantID:            uuid.NewString(),
				ReasonType:            "INVESTIGATION_IN_PROCESS",
				InvestigationMetadata: paymentModel.InvestigationPoPMetadata{},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error when updating",
			request: paymentModel.UpdatePaymentForInvestigationRequest{
				PaymentID:  uuid.NewString(),
				MerchantID: uuid.NewString(),
				ReasonType: "INVESTIGATION_IN_PROCESS",
				InvestigationMetadata: paymentModel.InvestigationPoPMetadata{
					Bucket: "test-bucket",
					Path:   "investigations/merchant-123/payment-456.png",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(false, errors.New("database connection error"))
			},
			wantErr: true,
		},
		{
			name: "SUCCESS: Update with INVESTIGATION_SUCCESS status",
			request: paymentModel.UpdatePaymentForInvestigationRequest{
				PaymentID:  uuid.NewString(),
				MerchantID: uuid.NewString(),
				ReasonType: "INVESTIGATION_SUCCESS",
				InvestigationMetadata: paymentModel.InvestigationPoPMetadata{
					Bucket:        "test-bucket",
					Path:          "investigations/merchant-123/payment-456.png",
					MerchantNotes: "Verified by ops team",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update with INVESTIGATION_FAILED status",
			request: paymentModel.UpdatePaymentForInvestigationRequest{
				PaymentID:  uuid.NewString(),
				MerchantID: uuid.NewString(),
				ReasonType: "INVESTIGATION_FAILED",
				InvestigationMetadata: paymentModel.InvestigationPoPMetadata{
					Bucket:        "test-bucket",
					Path:          "investigations/merchant-123/payment-456.png",
					MerchantNotes: "Investigation rejected",
				},
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.Anything,
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
					constant.StringMockType(),
				).Return(true, nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			ctx := context.WithValue(context.Background(), pdkConst.CtxSQLTableNameKey, "payments")
			err := repo.UpdatePaymentForInvestigation(ctx, tc.request)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestUpdatePaymentStatusWithReason(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	paymentID := "52442533-9366-4673-9ac5-4a1e7a071020"

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.MatchedBy(func(raw string) bool {
					return strings.Contains(raw, "status = ?") &&
						strings.Contains(raw, "reason_type = ?") &&
						strings.Contains(raw, "reason_description = ?") &&
						strings.Contains(raw, "WHERE uuid = ?;")
				}), mock.Anything, mock.Anything, mock.Anything, mock.Anything, paymentID).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.MatchedBy(func(raw string) bool {
					return strings.Contains(raw, "status = ?") &&
						strings.Contains(raw, "reason_type = ?") &&
						strings.Contains(raw, "reason_description = ?") &&
						strings.Contains(raw, "WHERE uuid = ?;")
				}), mock.Anything, mock.Anything, mock.Anything, mock.Anything, paymentID).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()
			assert.Equal(t, test.wantError, repo.UpdatePaymentStatusWithReason(t.Context(), paymentID, paymentModel.UpdatePaymentStatusWithReasonRequest{}))
		})
	}
}
