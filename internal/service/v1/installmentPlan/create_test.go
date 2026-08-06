package installmentplan_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	installmentplan "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/installmentPlan"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInstallmentPlanService_Create(t *testing.T) {
	merchantID := uuid.New().String()
	midID := uuid.New().String()

	midDetail := &creditcardCoreProcessorModel.MIDResponseData{
		Mid:              midID,
		Type:             "SETTLEMENT",
		TransactionType:  "INSTALLMENT",
		InstallmentTenor: 12,
		InstallmentType:  "INSTALLMENT",
	}

	validRequest := &installmentPlanModel.CreateInstallmentPlanRequest{
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "SETTLEMENT",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Test Plan",
		Description:    "Test Description",
		Tenor:          12,
		CardDetail: &installmentPlanModel.CardInstallmentPlanRequest{
			MidID:              midID,
			MidInstallmentType: midDetail.InstallmentType,
			AllowedBins:        []string{"123456", "654321"},
			Interest:           2.5,
		},
	}

	validRequestWithoutMerchant := &installmentPlanModel.CreateInstallmentPlanRequest{
		Acquirer:       "HARSYA",
		SettlementType: "SETTLEMENT",
		PaymentMethod:  constant.InstallmentPlanPaymentMethodCard,
		Title:          "Test Plan",
		Description:    "Test Description",
		Tenor:          12,
		CardDetail: &installmentPlanModel.CardInstallmentPlanRequest{
			MidID:       midID,
			AllowedBins: []string{"123456", "654321"},
			Interest:    2.5,
		},
	}

	validRequestNonCard := &installmentPlanModel.CreateInstallmentPlanRequest{
		MerchantID:     merchantID,
		Acquirer:       "HARSYA",
		SettlementType: "SETTLEMENT",
		PaymentMethod:  "OTHER",
		Title:          "Test Plan",
		Description:    "Test Description",
		Tenor:          12,
	}

	merchant := &merchantModel.Merchant{
		UUID: merchantID,
		Name: "Test Merchant",
	}

	testCases := []struct {
		name      string
		request   *installmentPlanModel.CreateInstallmentPlanRequest
		wantErr   bool
		mockSetup func(*repositoryMocks.IInstallmentPlanRepository, *serviceMocks.IMerchantService, *serviceMocks.ICreditCardService)
	}{
		{
			name:    "SUCCESS: Create installment plan with merchant",
			request: validRequest,
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Create installment plan without merchant",
			request: validRequestWithoutMerchant,
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "SUCCESS: Create installment plan with non-card payment method",
			request: validRequestNonCard,
			wantErr: false,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:    "ERROR: Merchant service error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Merchant not found",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Get MID detail error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Invalid settlement type",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				midDetailWrongType := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "DIRECT",
					TransactionType:  "INSTALLMENT",
					InstallmentTenor: 12,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailWrongType, nil)
			},
		},
		{
			name:    "ERROR: Direct pay transaction type",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				midDetailDirectPay := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "SETTLEMENT",
					TransactionType:  constant.CreditCardMidTransactionTypeDirectPay,
					InstallmentTenor: 12,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailDirectPay, nil)
			},
		},
		{
			name:    "ERROR: Mismatched installment tenor",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				midDetailWrongTenor := &creditcardCoreProcessorModel.MIDResponseData{
					Mid:              midID,
					Type:             "SETTLEMENT",
					TransactionType:  "INSTALLMENT",
					InstallmentTenor: 6,
				}
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetailWrongTenor, nil)
			},
		},
		{
			name:    "ERROR: Validate MID bins error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Repository create error",
			request: validRequest,
			wantErr: true,
			mockSetup: func(repo *repositoryMocks.IInstallmentPlanRepository, merchantSvc *serviceMocks.IMerchantService, ccSvc *serviceMocks.ICreditCardService) {
				merchantSvc.On("FindMerchantByID", mock.Anything, merchantID).Return(merchant, nil)
				ccSvc.On("GetMIDDetail", mock.Anything, midID).Return(midDetail, nil)
				ccSvc.On("ValidateMIDInstallmentBins", mock.Anything, mock.Anything).Return(nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockRepo := repositoryMocks.NewIInstallmentPlanRepository(t)
			mockMerchantSvc := serviceMocks.NewIMerchantService(t)
			mockCreditCardSvc := serviceMocks.NewICreditCardService(t)

			tc.mockSetup(mockRepo, mockMerchantSvc, mockCreditCardSvc)

			svc := installmentplan.NewInstallmentPlanService(mockLogger, mockRepo, mockCreditCardSvc, mockMerchantSvc)
			result, err := svc.Create(context.Background(), tc.request)

			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tc.request.Title, result.Title)
				assert.Equal(t, tc.request.Description, result.Description)
				assert.Equal(t, tc.request.Tenor, result.Tenor)
				assert.Equal(t, constant.InstallmentPlanStatusActive, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockMerchantSvc.AssertExpectations(t)
			mockCreditCardSvc.AssertExpectations(t)
		})
	}
}
