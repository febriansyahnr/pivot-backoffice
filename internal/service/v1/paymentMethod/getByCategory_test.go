package paymentMethodService

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestGetPaymentMethodByCategory(t *testing.T) {
	desc := "VA Bank Permata"
	logo := "https://sandbox.bca.co.id/api/images/logo-bca.png"
	bankName := "Permata Bank"

	paymentMethod := &paymentModel.PaymentMethod{
		UUID:        "e8a25416-050b-4c0c-9943-a258363b1c1f",
		Type:        "Virtual Account",
		Category:    "disbursement_top_up",
		Name:        "VA Permata",
		Description: &desc,
		Logo:        &logo,
		Acquirer:    "permata",
		BankName:    &bankName,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	paymentMethods := []*paymentModel.PaymentMethod{paymentMethod}

	tests := []struct {
		name       string
		category   string
		want       []*paymentModel.PaymentMethod
		mocksSetup func(

			paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
		)
		wantErr bool
	}{
		{
			name:     "SUCCESS: Get Payment Method By Category",
			category: paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			want:     paymentMethods,
			mocksSetup: func(

				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentMethodRepo.
					On(
						"GetAllPaymentMethodByCategory",
						constant.ValueCtxMockType(),
						mock.Anything).
					Return(paymentMethods, nil).
					Once()
			},
			wantErr: false,
		},
		{
			name:     "FAILED: Get Payment Method By Category",
			category: paymentConstant.PAYMENT_METHOD_CATEGORY_DISBURSEMENT_TOPUP,
			want:     paymentMethods,
			mocksSetup: func(

				paymentMethodRepo *repositoryMocks.IPaymentMethodRepository,
			) {
				paymentMethodRepo.
					On(
						"GetAllPaymentMethodByCategory",
						constant.ValueCtxMockType(),
						mock.Anything).
					Return(nil, assert.AnError).
					Once()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockPaymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)
			tt.mocksSetup(mockPaymentMethodRepo)

			svc := New(mockLogger, mockPaymentMethodRepo, snapCoreRepo, creditCardRepo)
			got, err := svc.FindPaymentMethodByCategory(context.Background(), tt.category)

			if tt.wantErr {
				assert.Empty(t, got)
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, got)
				assert.NoError(t, err)
			}

			mockPaymentMethodRepo.AssertExpectations(t)

		})
	}
}
