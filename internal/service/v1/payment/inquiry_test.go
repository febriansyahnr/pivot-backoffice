package paymentService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInquiryID(t *testing.T) {
	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService)
	}{
		{
			name:    "ERROR: get GetPaymentById repo",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: get GetPaymentById not found",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Inquiry EWallet Payment",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						UUID:   uuid.NewString(),
						Status: constant.UnifiedPaymentSessionStatusProcessing,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: constant.UnifiedPaymentMethodEWallet,
						},
					}, nil)

				unifiedPaymentSvc.On("InquiryEWalletPayment",
					mock.Anything, mock.Anything).
					Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "SUCCESS",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						UUID: uuid.NewString(),
					}, nil)
			},
		},
		{
			name:    "SUCCESS: Inquiry EWallet Payment",
			wantErr: false,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						UUID:   uuid.NewString(),
						Status: constant.UnifiedPaymentSessionStatusProcessing,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: constant.UnifiedPaymentMethodEWallet,
						},
					}, nil)

				unifiedPaymentSvc.On("InquiryEWalletPayment",
					mock.Anything, mock.Anything).
					Return(&paymentModel.Payment{
						UUID: uuid.NewString(),
					}, nil).Once()
			},
		},
		{
			name:    "ERROR: Inquiry Card Payment",
			wantErr: true,
			setupMock: func(paymentRepo *repositoryMocks.IPaymentRepository, unifiedPaymentSvc *serviceMocks.IUnifiedPaymentService) {
				paymentRepo.On("GetPaymentById",
					constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						UUID:   uuid.NewString(),
						Status: constant.UnifiedPaymentSessionStatusProcessing,
						PaymentMethod: paymentModel.PaymentMethod{
							Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
						},
					}, nil)

				unifiedPaymentSvc.On("InquiryCardPayment",
					mock.Anything, mock.Anything).
					Return(nil, constant.ErrSomeErrorForUnitTest).Once()
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentRepo := repositoryMocks.NewIPaymentRepository(t)
			unifiedPaymentSvc := serviceMocks.NewIUnifiedPaymentService(t)
			logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.setupMock(paymentRepo, unifiedPaymentSvc)
			paymentSvc := New(paymentRepo, logger, nil, nil, nil, nil, nil)
			WithUnifiedPaymentService(paymentSvc, unifiedPaymentSvc)

			ctx := context.Background()
			data, err := paymentSvc.InquiryPayment(ctx, &paymentModel.InquiryPaymentRequest{
				PaymentID: "payment-id",
			})
			if tc.wantErr {
				assert.Error(t, err)
				require.Empty(t, data)
			} else {
				assert.NoError(t, err)
				require.NotEmpty(t, data)
			}

			paymentRepo.AssertExpectations(t)
		})
	}
}
