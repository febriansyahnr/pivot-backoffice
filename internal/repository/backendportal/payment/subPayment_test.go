package paymentRepository

import (
	"context"
	"database/sql"
	"testing"

	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	mySqlExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetAutoSplitSubPayments(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	repo := New(db, mockLogger)

	mockResult := []*paymentModel.PaymentWithPaymentMethodDTO{
		{
			UUID:       "uuid-1",
			MerchantID: "merchant-1",
			Status:     "SETTLED",
			Type:       constant.UnifiedPaymentTypeSubPayment,
		},
		{
			UUID:       "uuid-2",
			MerchantID: "merchant-1",
			Status:     "PENDING",
			Type:       constant.UnifiedPaymentTypeSubPayment,
		},
	}

	emptyStr := ""
	wantResult := []*paymentModel.Payment{
		{
			UUID:       "uuid-1",
			MerchantID: "merchant-1",
			Status:     "SETTLED",
			Type:       constant.UnifiedPaymentTypeSubPayment,
			PaymentMethod: paymentModel.PaymentMethod{
				BankName: &emptyStr,
				Logo:     &emptyStr,
			},
		},
		{
			UUID:       "uuid-2",
			MerchantID: "merchant-1",
			Status:     "PENDING",
			Type:       constant.UnifiedPaymentTypeSubPayment,
			PaymentMethod: paymentModel.PaymentMethod{
				BankName: &emptyStr,
				Logo:     &emptyStr,
			},
		},
	}

	tests := []struct {
		name       string
		request    *paymentModel.GetSubPaymentsRequest
		setupMock  func()
		wantErr    string
		wantResult []*paymentModel.Payment
	}{
		{
			name: "when failed to get sub payments, then should return error",
			request: &paymentModel.GetSubPaymentsRequest{
				MerchantID: "merchant-1",
			},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Once().Return(constant.ErrSomeErrorForUnitTest)
			},
			wantErr: constant.ErrSomeErrorForUnitTest.Error(),
		},
		{
			name:    "when no filters provided, then should return sub payments with type filter only",
			request: &paymentModel.GetSubPaymentsRequest{},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Once().Return(nil)
			},
			wantResult: []*paymentModel.Payment{},
		},
		{
			name: "when merchant id filter provided, then should return filtered sub payments",
			request: &paymentModel.GetSubPaymentsRequest{
				MerchantID: "merchant-1",
			},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]*paymentModel.PaymentWithPaymentMethodDTO)) = mockResult
				}).Once().Return(nil)
			},
			wantResult: wantResult,
		},
		{
			name: "when status filter provided, then should return filtered sub payments with uppercased status",
			request: &paymentModel.GetSubPaymentsRequest{
				Status: "settled",
			},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), "SETTLED",
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]*paymentModel.PaymentWithPaymentMethodDTO)) = mockResult[:1]
				}).Once().Return(nil)
			},
			wantResult: wantResult[:1],
		},
		{
			name: "when reference id filter provided, then should return filtered sub payments",
			request: &paymentModel.GetSubPaymentsRequest{
				ReferenceID: "ref-123",
			},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(),
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]*paymentModel.PaymentWithPaymentMethodDTO)) = mockResult[:1]
				}).Once().Return(nil)
			},
			wantResult: wantResult[:1],
		},
		{
			name: "when all filters provided, then should return filtered sub payments",
			request: &paymentModel.GetSubPaymentsRequest{
				MerchantID:  "merchant-1",
				Status:      "settled",
				ReferenceID: "ref-123",
			},
			setupMock: func() {
				db.On(
					"SelectContext", constant.ValueCtxMockType(), mock.Anything, constant.StringMockType(), "SETTLED",
				).Run(func(args mock.Arguments) {
					(*args.Get(1).(*[]*paymentModel.PaymentWithPaymentMethodDTO)) = mockResult[:1]
				}).Once().Return(nil)
			},
			wantResult: wantResult[:1],
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			result, err := repo.GetAutoSplitSubPayments(context.Background(), tc.request)
			if tc.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, tc.wantResult, result)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
			}
		})
	}
}

func TestHardDeleteSubPayments(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID  = "c4f7d895-e7a7-4ffb-bf90-8064a1da2b27"
		referenceID = "8389a85a-f877-4e1d-b5cd-4d2ffca8cd2a"
	)

	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, merchantID, referenceID, constant.UnifiedPaymentTypeSubPayment).Once().Return(false, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("ExecContext", mock.Anything, mock.Anything, merchantID, referenceID, constant.UnifiedPaymentTypeSubPayment).Once().Return(true, nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			assert.Equal(t, tt.wantError, repo.HardDeleteAutoSplitSubPayments(t.Context(), merchantID, referenceID))

			db.AssertExpectations(t)
		})
	}
}

func TestGetSummaryAutoSplitPayment(t *testing.T) {
	db := mySqlExtMock.NewIMySqlExt(t)

	repo := New(db, nil)

	var (
		merchantID     = "merchant-id"
		referenceID    = "ref-id"
		maxCreatedDays = 14
	)

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *paymentModel.AutoSplitPaymentSummary
	}{
		{
			name: "ERROR: Some error", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID, maxCreatedDays,
				).Once().Return(assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID, maxCreatedDays,
				).Once().Return(nil)
			},
			wantResult: &paymentModel.AutoSplitPaymentSummary{},
		},
		{
			name: "ERROR: No rows found", // NOSONAR
			setupMock: func() {
				db.On(
					"GetContext", mock.Anything, mock.Anything, mock.Anything, merchantID, referenceID, maxCreatedDays,
				).Once().Return(sql.ErrNoRows)
			},
			wantResult: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			result, err := repo.GetSummaryAutoSplitPayment(t.Context(), &paymentModel.GetAutoSplitPaymentSummaryRequest{
				ReferenceID:     referenceID,
				MerchantID:      merchantID,
				MaxDateCreation: maxCreatedDays,
			})
			assert.Equal(t, tt.wantError, err)
			assert.Equal(t, tt.wantResult, result)

			db.AssertExpectations(t)
		})
	}
}
