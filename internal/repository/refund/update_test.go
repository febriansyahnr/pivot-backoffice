package refundRepository_test

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	refundModel "github.com/paper-indonesia/pivot-backoffice/internal/model/refund"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/refund"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateData(t *testing.T) {
	testCases := []struct {
		name        string
		refund      *refundModel.Refund
		mockDBError error
		expectError error
	}{
		{
			name: "success update",
			refund: &refundModel.Refund{
				UUID:            "uuid-123",
				Currency:        "IDR",
				Amount:          5000,
				Status:          "COMPLETED",
				Reason:          "Verified",
				Description:     "Refund success",
				DestinationType: "BANK_ACCOUNT",
				Method:          "TRANSFER",
			},
			mockDBError: nil,
			expectError: nil,
		},
		{
			name: "no rows affected, returns nil",
			refund: &refundModel.Refund{
				UUID: "uuid-456",
			},
			mockDBError: constant.ErrNoRowsAffected,
			expectError: nil,
		},
		{
			name: "db error",
			refund: &refundModel.Refund{
				UUID: "uuid-789",
			},
			mockDBError: errors.New("db error"),
			expectError: errors.New("db error"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			mockDB.On("ExecContext", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(false, tc.mockDBError)

			repo := New(mockDB, mockLogger)
			err := repo.UpdateData(context.Background(), tc.refund)

			if tc.expectError != nil {
				assert.EqualError(t, err, tc.expectError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockDB.AssertExpectations(t)
		})
	}
}
