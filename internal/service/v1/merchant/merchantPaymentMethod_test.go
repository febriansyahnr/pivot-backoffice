package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mockServices "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEnableAllPaymentMethod(t *testing.T) {
	mockPaymentMethod := mockRepositories.NewIPaymentMethodRepository(t)
	mockQrisSvc := mockServices.NewIQrisService(t)
	_, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)

	tests := []struct {
		name          string
		config        *config.Config
		merchant      *merchantModel.Merchant
		mockSetup     func()
		expectedError error
	}{
		{
			name: "should enable all payment methods successfully",
			config: &config.Config{Environment: constant.EnvironmentStaging, UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				MasterQRDuplicationID: "master-duplication-id",
			}},
			merchant: &merchantModel.Merchant{
				UUID: "merchant-uuid",
			},
			mockSetup: func() {
				mockPaymentMethod.On("GetAllPaymentMethodByCategory", constant.ValueCtxMockType(), constant.TypePayment).Return([]*paymentModel.PaymentMethod{
					{UUID: "pm-uuid-1", Type: paymentConstant.PAYMENT_METHOD_QRIS},
					{UUID: "pm-uuid-2", Type: "other"},
					{UUID: "pm-uuid-3", Type: paymentConstant.PAYMENT_METHOD_QRIS},
				}, nil).Once()
				mockQrisSvc.On("DuplicateRegistration", constant.ValueCtxMockType(), &qris.DuplicateRegistrationReq{
					SourceMerchantId: "master-duplication-id",
					TargetMerchantId: "merchant-uuid",
				}).Return("", nil).Once()
				mockPaymentMethod.On("UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), &paymentModel.PaymentMethodWithPivot{
					MerchantID:       "merchant-uuid",
					IsActive:         true,
					PaymentMethod:    paymentModel.PaymentMethod{UUID: "pm-uuid-1", Type: paymentConstant.PAYMENT_METHOD_QRIS},
					ChannelType:      constant.PaymentMethodChannelTypeAggregator,
					ActivationStatus: constant.PaymentMethodActivationStatusApproved,
				}).Return(nil).Once()
				mockPaymentMethod.On("UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), &paymentModel.PaymentMethodWithPivot{
					MerchantID:       "merchant-uuid",
					IsActive:         true,
					PaymentMethod:    paymentModel.PaymentMethod{UUID: "pm-uuid-2", Type: "other"},
					ChannelType:      constant.PaymentMethodChannelTypeAggregator,
					ActivationStatus: constant.PaymentMethodActivationStatusApproved,
				}).Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name:   "should return nil if environment is not staging",
			config: &config.Config{Environment: "production"},
			merchant: &merchantModel.Merchant{
				UUID: "merchant-uuid",
			},
			mockSetup:     func() {},
			expectedError: nil,
		},
		{
			name:   "should log error if GetAllPaymentMethodByCategory fails",
			config: &config.Config{Environment: constant.EnvironmentStaging},
			merchant: &merchantModel.Merchant{
				UUID: "merchant-uuid",
			},
			mockSetup: func() {
				mockPaymentMethod.On("GetAllPaymentMethodByCategory", constant.ValueCtxMockType(), constant.TypePayment).Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
			expectedError: constant.ErrSomeErrorForUnitTest,
		},
		{
			name: "should log error if DuplicateRegistration fails",
			config: &config.Config{Environment: constant.EnvironmentStaging, UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				MasterQRDuplicationID: "master-duplication-id",
			}},
			merchant: &merchantModel.Merchant{
				UUID: "merchant-uuid",
			},
			mockSetup: func() {
				mockPaymentMethod.On("GetAllPaymentMethodByCategory", constant.ValueCtxMockType(), constant.TypePayment).Return([]*paymentModel.PaymentMethod{
					{UUID: "pm-uuid-1", Type: paymentConstant.PAYMENT_METHOD_QRIS},
				}, nil).Once()
				mockQrisSvc.On("DuplicateRegistration", constant.ValueCtxMockType(), &qris.DuplicateRegistrationReq{
					SourceMerchantId: "master-duplication-id",
					TargetMerchantId: "merchant-uuid",
				}).Return("", errors.New("duplicate registration error")).Once()
			},
			expectedError: nil,
		},
		{
			name:   "should log error if UpsertPaymentMethodMerchantByIdAndMerchant fails",
			config: &config.Config{Environment: constant.EnvironmentStaging},
			merchant: &merchantModel.Merchant{
				UUID: "merchant-uuid",
			},
			mockSetup: func() {
				mockPaymentMethod.On("GetAllPaymentMethodByCategory", constant.ValueCtxMockType(), constant.TypePayment).Return([]*paymentModel.PaymentMethod{
					{UUID: "pm-uuid-1", Type: "other"},
				}, nil).Once()
				mockPaymentMethod.On("UpsertPaymentMethodMerchantByIdAndMerchant", constant.ValueCtxMockType(), mock.Anything).Return(errors.New("upsert error")).Once()
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mockSetup()

			service := &MerchantService{
				config:        tt.config,
				paymentMethod: mockPaymentMethod,
				qrisSvc:       mockQrisSvc,
				logger:        pdkLogger,
			}

			err = service.EnableAllPaymentMethod(ctx, tt.merchant)
			if tt.expectedError != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockPaymentMethod.AssertExpectations(t)
			mockQrisSvc.AssertExpectations(t)
		})
	}
}
