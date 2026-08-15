package paymentRepository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConst "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/mySqlExt"

	"github.com/google/uuid"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPaymentRepositoryGetInvestigatedPayments(t *testing.T) {
	testCase := []struct {
		name      string
		filter    *paymentModel.GetInvestigatedPaymentsFilterRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get investigated payments with default pagination",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:  1,
				Limit: 20,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with status filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:                1,
				Limit:               20,
				InvestigationStatus: paymentConst.InvestigationStatusInProcess,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with payment reference filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:               1,
				Limit:              20,
				PaymentReferenceID: uuid.NewString(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with merchant filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:       1,
				Limit:      20,
				MerchantID: uuid.NewString(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with date range filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:     1,
				Limit:    20,
				FromDate: func() *time.Time { t := time.Now().Add(-24 * time.Hour); return &t }(),
				ToDate:   func() *time.Time { t := time.Now(); return &t }(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with payment method filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:          1,
				Limit:         20,
				PaymentMethod: "QRIS",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with channel filter",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:    1,
				Limit:   20,
				Channel: "BCA",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get investigated payments with payment method and channel filters",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:          1,
				Limit:         20,
				PaymentMethod: "VIRTUAL_ACCOUNT",
				Channel:       "BCA",
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Limit defaults applied correctly",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:  0,
				Limit: 0,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Limit capped at 100",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:  1,
				Limit: 200,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error when selecting",
			filter: &paymentModel.GetInvestigatedPaymentsFilterRequest{
				Page:  1,
				Limit: 20,
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"SelectContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(errors.New("database error"))
				mysqlMock.On(
					"GetContext",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Once().Return(nil)
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			result, err := repo.GetInvestigatedPayments(ctx, tc.filter)

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestPaymentRepositoryUpdateInvestigationStatus(t *testing.T) {
	testCase := []struct {
		name      string
		request   paymentModel.UpdateInvestigationStatusRequest
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Update to INVESTIGATION_SUCCESS",
			request: paymentModel.UpdateInvestigationStatusRequest{
				PaymentID:   uuid.NewString(),
				Status:      paymentConst.InvestigationStatusSuccess,
				Notes:       func() *string { s := "Bank confirmed payment received"; return &s }(),
				CompletedAt: time.Now().UTC(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update to INVESTIGATION_FAILED",
			request: paymentModel.UpdateInvestigationStatusRequest{
				PaymentID:   uuid.NewString(),
				Status:      paymentConst.InvestigationStatusFailed,
				Notes:       func() *string { s := "Bank confirmed payment not received"; return &s }(),
				CompletedAt: time.Now().UTC(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Update without notes",
			request: paymentModel.UpdateInvestigationStatusRequest{
				PaymentID:   uuid.NewString(),
				Status:      paymentConst.InvestigationStatusSuccess,
				Notes:       nil,
				CompletedAt: time.Now().UTC(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Once().Return(true, nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: Database error when updating",
			request: paymentModel.UpdateInvestigationStatusRequest{
				PaymentID:   uuid.NewString(),
				Status:      paymentConst.InvestigationStatusSuccess,
				Notes:       nil,
				CompletedAt: time.Now().UTC(),
			},
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"ExecContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					constant.StringMockType(),
					constant.StringMockType(),
					mock.AnythingOfType("*string"),
					mock.AnythingOfType(constant.MockTypeTime),
					mock.AnythingOfType(constant.MockTypeTime),
					constant.StringMockType(),
				).Once().Return(false, errors.New("database connection error"))
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
			ctx := context.WithValue(context.Background(), mySqlExt.CtxSQLTableNameKey, "payments")
			err := repo.UpdateInvestigationStatus(ctx, tc.request)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)
		})
	}
}

func TestCalculateInvestigationMonthlyReconciliation(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	transaction := paymentModel.CalculateInvestigationMonthlyReconciliation{
		MerchantID:             "0fbb7e68-e08a-4b51-977d-336762c955b3",
		RawPaymentIDs:          []byte(`["08424aed-8cb6-465b-9c62-c8422b0778f7"]`),
		PaymentIDs:             []string{"08424aed-8cb6-465b-9c62-c8422b0778f7"},
		PaymentCount:           1,
		GrossAmount:            50_000,
		FeeAmount:              2_500,
		NetAmount:              47_500,
		PlatformLossPercentage: 10,
		PlatformMaxLoss:        100_000,
		PlatformLossAmount:     4_750,
		MerchantLossAmount:     42_750,
	}

	transaction2 := transaction
	transaction2.PlatformMaxLoss = 1_000
	transaction2.PlatformLossAmount = 1_000
	transaction2.MerchantLossAmount = 46_500

	transaction3 := transaction
	transaction3.PlatformLossPercentage = 0
	transaction3.PlatformMaxLoss = 0
	transaction3.PlatformLossAmount = 0
	transaction3.MerchantLossAmount = 47_500

	request := paymentModel.MonthlyReconciliationRequest{
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   time.Now(),
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult []paymentModel.CalculateInvestigationMonthlyReconciliation
	}{
		{
			name: "SUCCESS:Data not found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Return(sql.ErrNoRows)
			},
			wantError: nil,
		},
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:No rows", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Return(nil)
			},
			wantError:  nil,
			wantResult: []paymentModel.CalculateInvestigationMonthlyReconciliation{},
		},
		{
			name: "SUCCESS:Transaction found", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Run(func(args mock.Arguments) {
					trx := transaction
					trx.PlatformLossAmount, trx.MerchantLossAmount = 0, 0
					*args.Get(1).(*[]paymentModel.CalculateInvestigationMonthlyReconciliation) = []paymentModel.CalculateInvestigationMonthlyReconciliation{trx}
				}).Return(nil)
			},
			wantResult: []paymentModel.CalculateInvestigationMonthlyReconciliation{transaction},
		},
		{
			name: "SUCCESS:Transaction found (max loss exceeded)", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Run(func(args mock.Arguments) {
					trx := transaction2
					trx.PlatformLossAmount, trx.MerchantLossAmount = 0, 0
					*args.Get(1).(*[]paymentModel.CalculateInvestigationMonthlyReconciliation) = []paymentModel.CalculateInvestigationMonthlyReconciliation{trx}
				}).Return(nil)
			},
			wantResult: []paymentModel.CalculateInvestigationMonthlyReconciliation{transaction2},
		},
		{
			name: "SUCCESS:Transaction found (without compensation)", // NOSONAR
			setupMock: func() {
				db.On(
					"SelectContext", mock.Anything, mock.Anything, mock.Anything, paymentConst.InvestigationStatusFailed, request.StartDate, request.EndDate,
				).Once().Run(func(args mock.Arguments) {
					trx := transaction3
					trx.PlatformLossAmount, trx.MerchantLossAmount = 0, 0
					*args.Get(1).(*[]paymentModel.CalculateInvestigationMonthlyReconciliation) = []paymentModel.CalculateInvestigationMonthlyReconciliation{trx}
				}).Return(nil)
			},
			wantResult: []paymentModel.CalculateInvestigationMonthlyReconciliation{transaction3},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.CalculateInvestigationMonthlyReconciliation(t.Context(), request)

			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}

func TestInsertInvestigationMonthlyReconciliation(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"NamedExecContext", mock.Anything, mock.Anything, mock.Anything,
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.InsertInvestigationMonthlyReconciliation(t.Context(), paymentModel.PaymentInvestigationMonthlyReconciliation{}))
		})
	}
}

func TestUpdatePaymentInvestigationReconciliation(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)

	repo := New(db, nil)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "ABC",
				).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, "ABC",
				).Once().Return(true, nil)
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, repo.UpdatePaymentInvestigationReconciliation(t.Context(), paymentModel.PaymentInvestigationMonthlyReconciliation{
				PaymentIDs: []string{"ABC"},
			}))
		})
	}
}
