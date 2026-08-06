package settlementHoldService

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	settlementHold "github.com/paper-indonesia/pivot-backoffice/internal/model/settlementHolds"
	repoMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUpdate(t *testing.T) {
	now := time.Now().UTC()
	paymentID := uuid.NewString()
	merchantID := uuid.NewString()
	settlementHoldID := uuid.NewString()

	mockPayment := &paymentModel.Payment{
		UUID:       paymentID,
		MerchantID: merchantID,
		Amount:     decimal.NewFromInt(100000),
		Status:     constant.UnifiedPaymentSessionStatusPaid,
	}

	testCases := []struct {
		Name      string
		Input     *settlementHold.CreateUpdateSettlementHoldRequest
		MockSetup func(
			paymentSvc *serviceMock.IPaymentService,
			settlementSvc *serviceMock.ISettlementService,
			repo *repoMock.ISettlementHoldRepository,
		)
		WantErr     bool
		WantErrCode string
		Validate    func(t *testing.T, resp *settlementHold.CreateUpdateSettlementHoldResponse)
	}{
		{
			Name: "SUCCESS: Update existing settlement hold with different action",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "RELEASE",
				Reason:     "Payment verified",
				CreatedBy:  "user2",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				existingHold := &settlementHold.SettlementHold{
					UUID:       settlementHoldID,
					MerchantID: merchantID,
					PaymentID:  paymentID,
					Status:     "HOLD",
					CreatedBy:  "user1",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(existingHold, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(nil)
				repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Validate: func(t *testing.T, resp *settlementHold.CreateUpdateSettlementHoldResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, settlementHoldID, resp.UUID)
				assert.Equal(t, merchantID, resp.MerchantID)
				assert.Equal(t, paymentID, resp.PaymentID)
				assert.Equal(t, "RELEASE", resp.Status)
				assert.Equal(t, "Payment verified", resp.Reason)
			},
		},
		{
			Name: "SUCCESS: Update existing settlement hold with same action (no processing)",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "HOLD",
				Reason:     "Under review",
				CreatedBy:  "user2",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				existingHold := &settlementHold.SettlementHold{
					UUID:       settlementHoldID,
					MerchantID: merchantID,
					PaymentID:  paymentID,
					Status:     "HOLD",
					CreatedBy:  "user1",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(existingHold, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(nil)
				repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Validate: func(t *testing.T, resp *settlementHold.CreateUpdateSettlementHoldResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, "HOLD", resp.Status)
			},
		},
		{
			Name: "SUCCESS: Create new settlement hold",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "HOLD",
				Reason:     "Suspicious transaction",
				CreatedBy:  "user1",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(nil, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
			Validate: func(t *testing.T, resp *settlementHold.CreateUpdateSettlementHoldResponse) {
				assert.NotNil(t, resp)
				assert.Equal(t, merchantID, resp.MerchantID)
				assert.Equal(t, paymentID, resp.PaymentID)
				assert.Equal(t, "HOLD", resp.Status)
				assert.Equal(t, "Suspicious transaction", resp.Reason)
			},
		},
		{
			Name: "ERROR: Payment not found",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "HOLD",
				Reason:     "Test",
				CreatedBy:  "user1",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(nil, errors.New("payment not found"))
			},
			WantErr: true,
		},
		{
			Name: "ERROR: Get settlement hold by payment ID fails",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "HOLD",
				Reason:     "Test",
				CreatedBy:  "user1",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(nil, errors.New("database error"))
			},
			WantErr:     true,
			WantErrCode: constant.ErrValidateSettlementHold.Error(),
		},
		{
			Name: "ERROR: Process settlement hold/release fails",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "RELEASE",
				Reason:     "Test",
				CreatedBy:  "user2",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				existingHold := &settlementHold.SettlementHold{
					UUID:       settlementHoldID,
					MerchantID: merchantID,
					PaymentID:  paymentID,
					Status:     "HOLD",
					CreatedBy:  "user1",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(existingHold, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(errors.New("processing failed"))
			},
			WantErr:     true,
			WantErrCode: constant.ErrProcessSettlementHoldRelease.Error(),
		},
		{
			Name: "ERROR: Update settlement hold fails",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "RELEASE",
				Reason:     "Test",
				CreatedBy:  "user2",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				existingHold := &settlementHold.SettlementHold{
					UUID:       settlementHoldID,
					MerchantID: merchantID,
					PaymentID:  paymentID,
					Status:     "HOLD",
					CreatedBy:  "user1",
					CreatedAt:  now,
					UpdatedAt:  now,
				}
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(existingHold, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(nil)
				repo.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			WantErr:     true,
			WantErrCode: constant.ErrStoredSettlementHold.Error(),
		},
		{
			Name: "ERROR: Create settlement hold fails",
			Input: &settlementHold.CreateUpdateSettlementHoldRequest{
				MerchantID: merchantID,
				PaymentID:  paymentID,
				Action:     "HOLD",
				Reason:     "Test",
				CreatedBy:  "user1",
			},
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, settlementSvc *serviceMock.ISettlementService, repo *repoMock.ISettlementHoldRepository) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				repo.On("GetByPaymentID", mock.Anything, paymentID).Return(nil, nil)
				settlementSvc.On("ProcessSettlementHoldOrRelease", mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("create failed"))
			},
			WantErr:     true,
			WantErrCode: constant.ErrStoredSettlementHold.Error(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			paymentSvc := &serviceMock.IPaymentService{}
			settlementSvc := &serviceMock.ISettlementService{}
			repo := &repoMock.ISettlementHoldRepository{}
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.MockSetup(paymentSvc, settlementSvc, repo)

			svc := New(logger, repo, paymentSvc, settlementSvc, nil)
			resp, err := svc.CreateUpdate(context.Background(), tc.Input)

			if tc.WantErr {
				assert.NotNil(t, err)
				if tc.WantErrCode != "" {
					assert.Contains(t, err.Error(), tc.WantErrCode)
				}
			} else {
				assert.Nil(t, err)
				if tc.Validate != nil {
					tc.Validate(t, resp)
				}
			}

			paymentSvc.AssertExpectations(t)
			settlementSvc.AssertExpectations(t)
			repo.AssertExpectations(t)
		})
	}
}
