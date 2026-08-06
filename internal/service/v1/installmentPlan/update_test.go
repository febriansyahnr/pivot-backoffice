package installmentplan_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx/types"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	installmentplan "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/installmentPlan"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstallmentPlanService_Update(t *testing.T) {
	planUUID := uuid.New().String()
	merchantID := uuid.New().String()
	midID := uuid.New().String()
	now := time.Now().UTC()

	existingPlan := &installmentPlanModel.InstallmentPlan{
		UUID:           planUUID,
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "SETTLEMENT",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Old Plan",
		Description:    "Old Description",
		Tenor:          6,
		Status:         constant.InstallmentPlanStatusActive,
		Metadata:       types.JSONText(`{"card":{"midId":"old-mid","allowedBins":["123456"],"interest":1.5}}`),
		CreatedAt:      now,
		UpdatedAt:      now,
		PlanMetadata: &installmentPlanModel.InstallmentPlanMetadata{
			Card: &installmentPlanModel.CardInstallmentMetadata{
				MidID:       "old-mid",
				AllowedBins: []string{"123456"},
				Interest:    1.5,
			},
		},
	}

	tenor12 := 12
	interest25 := 2.5

	validUpdateRequest := &installmentPlanModel.UpdateInstallmentPlanRequest{
		UUID:           planUUID,
		MerchantID:     merchantID,
		Acquirer:       "MANDIRI",
		SettlementType: "DIRECT",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Updated Plan",
		Description:    "Updated Description",
		Tenor:          &tenor12,
		Status:         constant.InstallmentPlanStatusActive,
		CardDetail: &installmentPlanModel.UpdateCardInstallmentPlanRequest{
			MidID:       midID,
			AllowedBins: []string{"123456", "654321"},
			Interest:    &interest25,
		},
	}

	inactiveUpdateRequest := &installmentPlanModel.UpdateInstallmentPlanRequest{
		UUID:   planUUID,
		Status: constant.InstallmentPlanStatusInactive,
	}

	midDetail := &creditcardCoreProcessorModel.MIDResponseData{
		Mid:              midID,
		Type:             "DIRECT",
		TransactionType:  "INSTALLMENT",
		InstallmentTenor: 12,
	}

	testCases := []struct {
		name      string
		request   *installmentPlanModel.UpdateInstallmentPlanRequest
		wantErr   bool
		mockSetup func(*repositoryMocks.IInstallmentPlanRepository, *serviceMocks.ICreditCardService, *serviceMocks.IPaymentMethodService)
	}{
		{
			name:    "SUCCESS: Update installment plan with card validation",
			request: validUpdateRequest,
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*installmentPlanModel.InstallmentPlan")).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Update to inactive status (skip card validation)",
			request: inactiveUpdateRequest,
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*installmentPlanModel.InstallmentPlan")).Return(nil)
			},
		},
		{
			name:    "ERROR: Repository GetById error",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Installment plan not found",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(nil, nil)
			},
		},
		{
			name: "ERROR: Update validation error - empty bins for active plan",
			request: &installmentPlanModel.UpdateInstallmentPlanRequest{
				UUID:   planUUID,
				Status: constant.InstallmentPlanStatusActive,
				CardDetail: &installmentPlanModel.UpdateCardInstallmentPlanRequest{
					AllowedBins: []string{},
				},
			},
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
			},
		},
		{
			name:    "ERROR: Find installment plan in merchant payment methods",
			request: inactiveUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{}, errors.New("error"))
			},
		},
		{
			name:    "ERROR: There were installment plan in merchant payment methods",
			request: inactiveUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				paymentMethodSvc.On("GetPaymentMethodByMerchant", mock.Anything, mock.Anything).Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							UUID: uuid.NewString(),
						},
						MerchantID: uuid.NewString(),
						IsActive:   true,
					},
				}, nil)
			},
		},
		{
			name:    "ERROR: Get MID detail error",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Invalid settlement type",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				midDetailWrongType := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "SETTLEMENT",
					TransactionType:  "INSTALLMENT",
					InstallmentTenor: 12,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailWrongType, nil)
			},
		},
		{
			name:    "ERROR: Direct pay transaction type",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				midDetailDirectPay := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "DIRECT",
					TransactionType:  constant.CreditCardMidTransactionTypeDirectPay,
					InstallmentTenor: 12,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailDirectPay, nil)
			},
		},
		{
			name:    "ERROR: Mismatched installment tenor",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				midDetailWrongTenor := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "DIRECT",
					TransactionType:  "INSTALLMENT",
					InstallmentTenor: 6,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailWrongTenor, nil)
			},
		},
		{
			name:    "ERROR: Validate MID bins error",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Repository update error",
			request: validUpdateRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, ccSvc *serviceMocks.ICreditCardService, paymentMethodSvc *serviceMocks.IPaymentMethodService) {
				repo.On("GetById", mock.Anything, planUUID).Return(existingPlan, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(nil)
				repo.On("Update", mock.Anything, mock.AnythingOfType("*installmentPlanModel.InstallmentPlan")).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIInstallmentPlanRepository(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)
			mockPaymentMethodSvc := serviceMocks.NewIPaymentMethodService(t)

			tc.mockSetup(mockRepo, mockCreditCardSvc, mockPaymentMethodSvc)

			svc := installmentplan.NewInstallmentPlanService(mockLogger, mockRepo, mockCreditCardSvc, nil)
			installmentplan.WithPaymentMethodService(svc, mockPaymentMethodSvc)
			result, err := svc.Update(context.Background(), tc.request)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)

				// Verify updated fields
				if tc.request.Title != "" {
					assert.Equal(t, tc.request.Title, result.Title)
				}
				if tc.request.Description != "" {
					assert.Equal(t, tc.request.Description, result.Description)
				}
				if tc.request.Tenor != nil {
					assert.Equal(t, *tc.request.Tenor, result.Tenor)
				}
				if tc.request.Status != "" {
					assert.Equal(t, tc.request.Status, result.Status)
				}
			}

			mockRepo.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
		})
	}
}
