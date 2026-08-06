package settlementHoldService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	repoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetSettlementHoldByPaymentID(t *testing.T) {
	now := time.Now().UTC()
	paymentID := uuid.NewString()
	merchantID := uuid.NewString()
	settlementHoldID := uuid.NewString()

	mockSettlementHold := &settlementHold.SettlementHold{
		UUID:       settlementHoldID,
		MerchantID: merchantID,
		PaymentID:  paymentID,
		Status:     "HOLD",
		CreatedBy:  "user1",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	testCases := []struct {
		Name      string
		PaymentID string
		MockSetup func(repo *repoMock.ISettlementHoldRepository)
		WantErr   bool
		Validate  func(t *testing.T, result *settlementHold.SettlementHold)
	}{
		{
			Name:      "SUCCESS: Get settlement hold by payment ID",
			PaymentID: paymentID,
			MockSetup: func(repo *repoMock.ISettlementHoldRepository) {
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(mockSettlementHold, nil)
			},
			WantErr: false,
			Validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.NotNil(t, result)
				assert.Equal(t, settlementHoldID, result.UUID)
				assert.Equal(t, merchantID, result.MerchantID)
				assert.Equal(t, paymentID, result.PaymentID)
				assert.Equal(t, "HOLD", result.Status)
				assert.Equal(t, "user1", result.CreatedBy)
			},
		},
		{
			Name:      "SUCCESS: Settlement hold not found returns nil",
			PaymentID: paymentID,
			MockSetup: func(repo *repoMock.ISettlementHoldRepository) {
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(nil, nil)
			},
			WantErr: false,
			Validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.Nil(t, result)
			},
		},
		{
			Name:      "ERROR: Repository error",
			PaymentID: paymentID,
			MockSetup: func(repo *repoMock.ISettlementHoldRepository) {
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(nil, errors.New("database error"))
			},
			WantErr: true,
			Validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.Nil(t, result)
			},
		},
		{
			Name:      "ERROR: Empty payment ID",
			PaymentID: "",
			MockSetup: func(repo *repoMock.ISettlementHoldRepository) {
				repo.On("GetByPaymentID", mock.Anything, "").Return(nil, errors.New("invalid payment id"))
			},
			WantErr: true,
			Validate: func(t *testing.T, result *settlementHold.SettlementHold) {
				assert.Nil(t, result)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			repo := &repoMock.ISettlementHoldRepository{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.MockSetup(repo)

			svc := New(logger, repo, nil, nil, nil)
			result, err := svc.GetSettlementHoldByPaymentID(context.Background(), tc.PaymentID)

			if tc.WantErr {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), constant.ErrGetSettlementHold.Error())
			} else {
				assert.Nil(t, err)
			}

			if tc.Validate != nil {
				tc.Validate(t, result)
			}

			repo.AssertExpectations(t)
		})
	}
}