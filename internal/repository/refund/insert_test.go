package refundRepository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/refund"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRefundRepository_Insert(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name         string
		refund       *refundModel.Refund
		mockDBReturn bool
		mockDBError  error
		expectError  error
	}{
		{
			name: "success insert",
			refund: &refundModel.Refund{
				UUID:              "uuid-123",
				MerchantID:        "merchant-123",
				ClientReferenceID: "client-ref-123",
				PaymentID:         "payment-123",
				PaymentChargeID:   "charge-123",
				Currency:          "IDR",
				Amount:            10000,
				Status:            "PENDING",
				Reason:            "Customer Request",
				Description:       "Refund for order #123",
				DestinationType:   "BANK_ACCOUNT",
				Method:            "TRANSFER",
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			mockDBReturn: true,
			mockDBError:  nil,
			expectError:  nil,
		},
		{
			name: "db error",
			refund: &refundModel.Refund{
				UUID:              "uuid-456",
				MerchantID:        "merchant-456",
				ClientReferenceID: "client-ref-456",
			},
			mockDBReturn: false,
			mockDBError:  errors.New("db error"),
			expectError:  errors.New("db error"),
		},
		{
			name: "no rows affected",
			refund: &refundModel.Refund{
				UUID:              "uuid-789",
				MerchantID:        "merchant-789",
				ClientReferenceID: "client-ref-789",
			},
			mockDBReturn: false,
			mockDBError:  nil,
			expectError:  constant.ErrNoRowsAffected,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			mockDB.On("NamedExecContext", mock.Anything, mock.Anything, tc.refund).
				Return(tc.mockDBReturn, tc.mockDBError)

			repo := New(mockDB, mockLogger)
			err := repo.Insert(context.Background(), tc.refund)

			if tc.expectError != nil {
				assert.EqualError(t, err, tc.expectError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
