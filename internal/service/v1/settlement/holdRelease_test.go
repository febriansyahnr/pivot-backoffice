package settlementService_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	orchestrator_model "github.com/paper-indonesia/pivot-backoffice/internal/model/orchestrator"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	settlementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/settlement"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlement"
	settlementService "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/settlement"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	logger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProcessSettlementHoldOrRelease(t *testing.T) {
	log := logger.NewSlogger(logger.Config{})

	paymentID := uuid.NewString()
	mockPayment := &paymentModel.Payment{
		UUID:   paymentID,
		Amount: decimal.NewFromInt(100000),
	}

	testCases := []struct {
		Name           string
		PaymentID      string
		Action         string
		LastActionTime time.Time
		MockSetup      func(
			paymentSvc *serviceMock.IPaymentService,
			accountTxRepo *repositoryMock.IAccountTransactionRepository,
			settlementSvc *serviceMock.ISettlementService,
		)
		WantErr bool
	}{
		{
			Name:      "SUCCESS: Hold action sets isOnHold to true",
			PaymentID: paymentID,
			Action:    constant.SettlementHoldActionHold,
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, true).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:           "SUCCESS: Release action sets isOnHold to false",
			PaymentID:      paymentID,
			Action:         constant.SettlementHoldActionRelease,
			LastActionTime: time.Now().UTC(),
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, false).Return(nil)
				accountTxRepo.On("GetPastDueSettlementTransactions", mock.Anything, mock.Anything).Return([]*orchestrator_model.AccountTransaction{
					{
						UUID: uuid.New(),
						Type: constant.TypeFee,
					},
					{
						UUID: uuid.New(),
						Type: constant.TypePayment,
					},
				}, nil)
				settlementSvc.On("ProcessSettlementTransactionFee", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				settlementSvc.On("ProcessSettlement", mock.Anything, mock.Anything).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:      "SUCCESS: Release action sets isOnHold to false",
			PaymentID: paymentID,
			Action:    constant.SettlementHoldActionRelease,
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, false).Return(nil)
			},
			WantErr: false,
		},
		{
			Name:      "ERROR: Payment not found",
			PaymentID: paymentID,
			Action:    constant.SettlementHoldActionHold,
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(nil, errors.New("payment not found"))
			},
			WantErr: true,
		},
		{
			Name:      "ERROR: Update settlement hold fails",
			PaymentID: paymentID,
			Action:    constant.SettlementHoldActionHold,
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, true).Return(errors.New("database error"))
			},
			WantErr: true,
		},
		{
			Name:           "ERROR: Release action",
			PaymentID:      paymentID,
			Action:         constant.SettlementHoldActionRelease,
			LastActionTime: time.Now().UTC(),
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, false).Return(nil)
				accountTxRepo.On("GetPastDueSettlementTransactions", mock.Anything, mock.Anything).Return([]*orchestrator_model.AccountTransaction{
					{
						UUID: uuid.New(),
						Type: constant.TypeFee,
					},
					{
						UUID: uuid.New(),
						Type: constant.TypePayment,
					},
				}, nil)
				settlementSvc.On("ProcessSettlementTransactionFee", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
				settlementSvc.On("ProcessSettlement", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			},
			WantErr: true,
		},
		{
			Name:      "SUCCESS: Unknown action defaults to hold (isOnHold = false)",
			PaymentID: paymentID,
			Action:    "UNKNOWN_ACTION",
			MockSetup: func(paymentSvc *serviceMock.IPaymentService, accountTxRepo *repositoryMock.IAccountTransactionRepository, settlementSvc *serviceMock.ISettlementService) {
				paymentSvc.On("GetDetailByID", mock.Anything, paymentID).Return(mockPayment, nil)
				accountTxRepo.On("UpdateSettlementHoldByReferenceID", mock.Anything, paymentID, false).Return(nil)
			},
			WantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			paymentSvc := &serviceMock.IPaymentService{}
			accountTxRepo := &repositoryMock.IAccountTransactionRepository{}
			settlementSvc := serviceMock.NewISettlementService(t)

			tc.MockSetup(paymentSvc, accountTxRepo, settlementSvc)

			svc := New(log, accountTxRepo, WithPaymentSvc(paymentSvc))
			settlementService.WithInternalSvc(svc, settlementSvc)
			err := svc.ProcessSettlementHoldOrRelease(context.Background(), &settlementModel.ProcessHoldReleaseSettlementRequest{
				ReferenceID:    tc.PaymentID,
				Action:         tc.Action,
				LastActionTime: tc.LastActionTime,
			})

			if tc.WantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}

			paymentSvc.AssertExpectations(t)
			accountTxRepo.AssertExpectations(t)
		})
	}
}
