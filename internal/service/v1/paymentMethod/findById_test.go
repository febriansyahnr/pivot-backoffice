package paymentMethodService

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFindPaymentMethodByIdAndMerchant(t *testing.T) {
	repo := repositoryMocks.NewIPaymentMethodRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
	creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)

	validPaymentMethodID := uuid.NewString()
	validMerchantID := uuid.NewString()
	tests := []struct {
		name         string
		modifierMock func()
		wantErr      bool
	}{
		{
			name: "ERROR: FindPaymentMethodByIdAndMerchant error",
			modifierMock: func() {
				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(nil, constant.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
		{
			name: "ERROR: FindPaymentMethodByIdAndMerchant not found",
			modifierMock: func() {
				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "SUCCESS",
			modifierMock: func() {
				repo.On(
					"FindPaymentMethodByIdAndMerchant",
					constant.ValueCtxMockType(),
					constant.StringMockType(),
					constant.StringMockType(),
				).Once().Return(&paymentModel.PaymentMethodWithPivot{
					MerchantID: validMerchantID,
				}, nil)
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.modifierMock()

			svc := New(logger, repo, snapCoreRepo, creditCardRepo)
			got, err := svc.FindPaymentMethodByIdAndMerchant(context.Background(), validPaymentMethodID, validMerchantID)

			if tt.wantErr {
				assert.Empty(t, got)
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, got)
				assert.NoError(t, err)
			}
		})
	}
}
