package settlementHold

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetByPaymentID(t *testing.T) {
	now := time.Now().UTC()
	paymentID := "payment-123"

	expectedData := &settlementHold.SettlementHold{
		UUID:       "uuid-123",
		MerchantID: "merchant-123",
		PaymentID:  paymentID,
		Status:     "HOLD",
		CreatedBy:  "admin",
		CreatedAt:  now,
		UpdatedAt:  now,
		DeletedAt:  sql.NullTime{},
	}

	testCases := []struct {
		name      string
		paymentID string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt)
		wantErr   bool
		validate  func(t *testing.T, result *settlementHold.SettlementHold)
	}{
		{
			name:      "SUCCESS: Get settlement hold by payment ID",
			paymentID: paymentID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*settlementHold.SettlementHold"),
					mock.AnythingOfType("string"),
					paymentID,
				).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*settlementHold.SettlementHold)
					*arg = *expectedData
				}).Return(nil)
			},
			wantErr: false,
			validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.NotNil(t, result)
				assert.Equal(t, expectedData.UUID, result.UUID)
				assert.Equal(t, expectedData.PaymentID, result.PaymentID)
				assert.Equal(t, expectedData.Status, result.Status)
			},
		},
		{
			name:      "ERROR: Settlement hold not found",
			paymentID: "non-existent",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*settlementHold.SettlementHold"),
					mock.AnythingOfType("string"),
					"non-existent",
				).Return(sql.ErrNoRows)
			},
			wantErr: false,
			validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.Nil(t, result)
			},
		},
		{
			name:      "ERROR: Database query fails",
			paymentID: paymentID,
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt) {
				mysqlMock.On(
					"GetContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("*settlementHold.SettlementHold"),
					mock.AnythingOfType("string"),
					paymentID,
				).Return(errors.New("database error"))
			},
			wantErr: true,
			validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysql := mysqlMocks.NewIMySqlExt(t)

			tc.mockSetup(mockMysql)

			repo := New(mockMysql, mockLogger)
			result, err := repo.GetByPaymentID(context.Background(), tc.paymentID)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			tc.validate(t, result)
			mockMysql.AssertExpectations(t)
		})
	}
}
