package paymentRepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	constant "github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	pdkLogMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"

	"github.com/jmoiron/sqlx/types"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetTodayPaymentStatusInsight(t *testing.T) {
	var (
		ctx           = context.Background()
		mockDB        = mysqlMocks.IMySqlExt{}
		mockLogger, _ = loggerMocks.NewZapLogger(loggerMocks.Config{})
		paymentRepo   = PaymentRepository{
			db:     &mockDB,
			logger: mockLogger,
		}

		loc, _     = time.LoadLocation(constant.TimeLoc)
		now        = time.Now().In(loc)
		startOfDay = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).UTC()
	)
	testCases := []struct {
		name      string
		payload   paymentModel.PaymentInsightOption
		callMock  func(t *testing.T)
		want      paymentModel.PaymentInsightItem
		wantErr   error
		shouldErr bool
	}{
		{
			name: "when merchant payment exist, then should return correct data",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: "valid-merchant-id",
				Status:     constantPayment.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func(t *testing.T) {
				mockDB.Mock.Test(t)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, "valid-merchant-id", startOfDay, constantPayment.PAYMENT_STATUS_SUCCESS).
					Return(nil).Run(func(args mock.Arguments) {
					*args.Get(1).(*paymentModel.PaymentInsightQueryResult) = paymentModel.PaymentInsightQueryResult{
						Total:       2,
						TotalAmount: float64(100000),
					}
				}).Once()
			},
			want: paymentModel.PaymentInsightItem{
				Total: 2,
				TotalAmount: commonModel.Amount{
					Currency: "",
					Value:    "100000.00",
				},
			},
		},
		{
			name: "when database was down, then should return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: "invalid-merchant-id",
				Status:     constantPayment.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func(t *testing.T) {
				mockDB.Mock.Test(t)

				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, "invalid-merchant-id", startOfDay, constantPayment.PAYMENT_STATUS_SUCCESS).
					Return(errors.New("database down"))

			},
			shouldErr: true,
			wantErr:   errors.New("database down"),
		},
		{
			name: "when no merchant payment exist, then should not return error",
			payload: paymentModel.PaymentInsightOption{
				MerchantID: "valid-merchant-id",
				Status:     constantPayment.PAYMENT_STATUS_SUCCESS,
			},
			callMock: func(t *testing.T) {
				mockDB.Mock.Test(t)
				mockDB.On("GetContext", mock.Anything, mock.Anything, mock.Anything, "valid-merchant-id", startOfDay, constantPayment.PAYMENT_STATUS_SUCCESS).
					Return(sql.ErrNoRows)

			},
			want: paymentModel.PaymentInsightItem{
				Total: 0,
				TotalAmount: commonModel.Amount{
					Currency: "",
					Value:    "0.00",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock(t)

			insight, err := paymentRepo.GetTodayPaymentStatusInsight(ctx, tc.payload)

			if tc.shouldErr {
				assert.NotNil(t, err)
				assert.Equal(t, tc.wantErr, err)
				return
			}

			assert.Nil(t, err)
			assert.Equal(t, tc.want.Total, insight.Total)
			assert.Equal(t, tc.want.TotalAmount.Value, insight.TotalAmount.Value)
		})
	}
}

func TestGetPaymentDashboardInsights(t *testing.T) {
	db := mysqlMocks.NewIMySqlExt(t)
	logger := pdkLogMock.NewILogger(t)

	repo := New(db, logger)

	args := []any{
		mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	}
	failureReasonsText := `[{"count": 1, "percentage": 100.00, "failureCode": "AUTHENTICATION_FAILED"}]`

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *paymentModel.PaymentDashboardInsights
	}{
		{
			name: "ERROR:Executing aggregate query", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Return(assert.AnError)
				logger.On("Error", mock.Anything, "failed while executing aggregate query", mock.Anything).Once().Return()
			},
			wantError: assert.AnError,
		},
		{
			name: "ERROR:Unmarshal failure reasons", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*paymentModel.PaymentDashboardInsights) = paymentModel.PaymentDashboardInsights{
						RawFailureReasons: types.NullJSONText{
							Valid:    true,
							JSONText: []byte("B"),
						},
					}
				}).Return(nil)
				logger.On("Error", mock.Anything, "failed when unmarshal failure reasons", mock.Anything).Once().Return()
			},
			wantError: errors.New("json unmarshal: invalid character 'B' looking for beginning of value"),
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				db.On("GetContext", args...).Once().Run(func(args mock.Arguments) {
					*args.Get(1).(*paymentModel.PaymentDashboardInsights) = paymentModel.PaymentDashboardInsights{
						PaidCount:     5,
						RefundedCount: 2,
						FailedCount:   1,
						FailedTotal:   10_000,
						RawFailureReasons: types.NullJSONText{
							Valid:    true,
							JSONText: []byte(failureReasonsText),
						},
					}
				}).Return(nil)
			},
			wantResult: &paymentModel.PaymentDashboardInsights{
				PaidCount:     5,      // NOSONAR
				RefundedCount: 2,      // NOSONAR
				FailedCount:   1,      // NOSONAR
				FailedTotal:   10_000, // NOSONAR
				RawFailureReasons: types.NullJSONText{
					Valid:    true,
					JSONText: []byte(failureReasonsText),
				},
				FailureReasons: []paymentModel.PaymentDashboardInsightFailureReason{
					{
						Count:       1,                       // NOSONAR
						Percentage:  json.Number("100.00"),   // NOSONAR
						FailureCode: "AUTHENTICATION_FAILED", // NOSONAR
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := repo.GetPaymentDashboardInsights(t.Context(), paymentModel.GetPaymentDashboardInsightRequest{})
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
