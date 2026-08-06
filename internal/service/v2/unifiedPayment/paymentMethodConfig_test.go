package unifiedPaymentService_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	installmentPlanModel "github.com/paper-indonesia/pivot-backoffice/internal/model/installmentPlan"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	repositoryMock "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetPaymentMethodConfig(t *testing.T) {
	minAmount := 10000.00
	maxAmount := 1000000000.00
	cfg := &config.Config{
		Environment: "test",
		UnifiedPaymentConfig: config.UnifiedPaymentConfig{
			VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
				MinAmount:             &minAmount,
				MaxAmount:             &maxAmount,
				MaxExpiryDuration:     10,
				MaxExpiryDurationUnit: "HOUR",
			},
			QrConfig: &config.UnifiedPaymentQrConfig{
				MinAmount:             &minAmount,
				MaxAmount:             &maxAmount,
				MaxExpiryDuration:     10,
				MaxExpiryDurationUnit: "HOUR",
			},
			CardConfig: &config.UnifiedPaymentCardConfig{
				MinAmount:             &minAmount,
				MaxAmount:             &maxAmount,
				AcceptedChannels:      []string{"MASTERCARD", "VISA"},
				MaxExpiryDuration:     30,
				MaxExpiryDurationUnit: "DAYS",
			},
			EwalletConfig: &config.UnifiedPaymentEwalletConfig{
				MinAmount:             &minAmount,
				MaxAmount:             &maxAmount,
				MaxExpiryDuration:     30,
				MaxExpiryDurationUnit: "MINUTES",
			},
		},
	}
	log, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	paymentRepo := repositoryMock.NewIPaymentRepository(t)
	paymentMethodRepo := repositoryMock.NewIPaymentMethodRepository(t)
	accountTrxRepo := repositoryMock.NewIAccountTransactionRepository(t)
	installmentPlanSvc := serviceMock.NewIInstallmentPlanService(t)
	merchantRepo := repositoryMock.NewIMerchantRepository(t)

	// KYC-approved merchant: merchantId is used as-is.
	kycApprovedMerchant := &merchantModel.Merchant{
		KYCStatus: sql.NullString{
			String: c.KYCStatusApproved,
			Valid:  true,
		},
	}
	// Non-KYC merchant: merchantId is overridden with ParentID before fetching payment methods.
	nonKycMerchant := &merchantModel.Merchant{
		KYCStatus: sql.NullString{
			String: "NON_KYC",
			Valid:  true,
		},
		ParentID: sql.NullString{
			String: "parent-merchant-id",
			Valid:  true,
		},
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Failed to get merchant data",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Merchant not found",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Get list payment method",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(kycApprovedMerchant, nil)

				paymentMethodRepo.On("GetListPaymentMethodMerchant",
					c.ValueCtxMockType(),
					c.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get Installment Plans",
			wantErr: true,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(kycApprovedMerchant, nil)

				paymentMethodRepo.On("GetListPaymentMethodMerchant",
					c.ValueCtxMockType(),
					c.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						},
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
									InstallmentPlanIDs: []string{uuid.NewString()},
								},
							},
						},
						IsActive: true,
					},
				}, nil)

				installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return(
					nil, int64(0), errors.New("error"),
				).Once()
			},
		},
		{
			name:    "SUCCESS: Non-KYC merchant uses parent id",
			wantErr: false,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(nonKycMerchant, nil)

				paymentMethodRepo.On("GetListPaymentMethodMerchant",
					c.ValueCtxMockType(),
					c.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return([]*paymentModel.PaymentMethodWithPivot{}, nil)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func() {
				merchantRepo.On("FindMerchantByID",
					c.ValueCtxMockType(),
					mock.AnythingOfType("string"),
				).Once().Return(kycApprovedMerchant, nil)

				paymentMethodRepo.On("GetListPaymentMethodMerchant",
					c.ValueCtxMockType(),
					c.PtrGetPaymentMethodFilterRequestMockType(),
				).Once().Return([]*paymentModel.PaymentMethodWithPivot{
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
						},
						IsActive: true,
					},
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_QRIS,
						},
						IsActive: true,
					},
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						},
						IsActive: true,
					},
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_EWALLET,
						},
						IsActive: true,
					},
					{
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_INSTALLMENT,
						},
						MerchantConfigObj: &paymentModel.PaymentMethodMerchantConfigObject{
							PartnerConfig: &paymentMethodModel.SetupPaymentMethodPartnerConfigRequest{
								Installment: &paymentMethodModel.SetupPaymentMethodPartnerConfigForInstallmentRequest{
									InstallmentPlanIDs: []string{uuid.NewString()},
								},
							},
						},
						IsActive: true,
					},
				}, nil)

				installmentPlanSvc.On("List", mock.Anything, mock.Anything).Return(
					[]*installmentPlanModel.InstallmentPlan{
						{
							UUID:  uuid.NewString(),
							Title: "Program Title",
							Tenor: 3,
							PlanMetadata: &installmentPlanModel.InstallmentPlanMetadata{
								Card: &installmentPlanModel.CardInstallmentMetadata{
									AllowedBins:   []string{"123456"},
									MinimumAmount: 10000,
									MaximumAmount: 1000000,
									Interest:      1,
								},
							},
						},
					}, int64(1), nil,
				).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			svc := New(cfg, log, paymentRepo, paymentMethodRepo, accountTrxRepo,
				WithMerchantRepo(merchantRepo),
			)
			WithInstallmentPlanService(svc, installmentPlanSvc)
			_, err := svc.GetPaymentMethodConfig(context.Background(), uuid.NewString())
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
