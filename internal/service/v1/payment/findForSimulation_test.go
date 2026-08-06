package paymentService

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	constantPayment "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestFindPaymentForSimulationByID(t *testing.T) {
	paymentRepo := repositoryMocks.NewIPaymentRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	merchantRepo := repositoryMocks.NewIMerchantRepository(t)
	paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)

	validReferenceID := "mock-reference-id"
	paymentMetadata := map[string]any{
		"snapCore": "OK",
	}

	testCases := []struct {
		name      string
		wantErr   bool
		setupMock func()
	}{
		{
			name:    "ERROR: Get payment by ID error repo",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get payment by ID not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "ERROR: Get payment method by ID not found",
			wantErr: true,
			setupMock: func() {
				paymentRepo.On("GetPaymentById", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&paymentModel.Payment{
						PaymentMethodID: "mock-payment-method",
						MerchantID:      "mock-merchant-id",
						ReferenceID:     &validReferenceID,
						Metadata:        &paymentMetadata,
					}, nil)

				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrPaymentMethodNotFound)
			},
		},
		{
			name:    "ERROR: Get payment method by ID error repo",
			wantErr: true,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get merchant ID error repo",
			wantErr: true,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
		{
			name:    "ERROR: Get merchant ID not found",
			wantErr: true,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, nil)
			},
		},
		{
			name:    "SUCCESS: For payment method QRIS",
			wantErr: false,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_QRIS}, nil)

				merchantRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).
					Return(&merchantModel.Merchant{Name: "Mock Merchant Name"}, nil)
			},
		},
		{
			name:    "ERROR: Get payment for payment method VA",
			wantErr: true,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Twice().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT}, nil)

				paymentRepo.On("GetPaymentItemsByPaymentId", constant.ValueCtxMockType(), constant.StringMockType()).
					Once().Return(nil, constant.ErrSomeErrorForUnitTest)

			},
		},
		{
			name:    "SUCCESS: For payment method VA",
			wantErr: false,
			setupMock: func() {
				paymentMethodRepo.On("GetPaymentMethodById", constant.ValueCtxMockType(), constant.StringMockType()).
					Twice().Return(&paymentModel.PaymentMethod{Type: constantPayment.PAYMENT_METHOD_VIRTUAL_ACCOUNT}, nil)

				paymentRepo.On("GetPaymentItemsByPaymentId", constant.ValueCtxMockType(), constant.StringMockType()).
					Return([]*paymentModel.PaymentItem{
						{
							UUID: "mock-payment-item-id",
						},
					}, nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			tc.setupMock()

			paymentSvc := New(paymentRepo, logger, nil, nil, merchantRepo, paymentMethodRepo, nil)

			ctx := context.Background()
			result, err := paymentSvc.FindPaymentForSimulationByID(ctx, "mock-id")

			if tc.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			paymentRepo.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			paymentMethodRepo.AssertExpectations(t)
		})
	}
}
