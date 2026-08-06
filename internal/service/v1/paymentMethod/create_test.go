package paymentMethodService

import (
	"context"
	"testing"

	"github.com/google/uuid"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentMethodModel "github.com/paper-indonesia/pivot-backoffice/internal/model/paymentMethod"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	validUUID := uuid.NewString()

	testCases := []struct {
		name       string
		input      *paymentMethodModel.CreatePaymentMethodRequest
		mocksSetup func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository)
		wantErr    bool
	}{
		{
			name: "SUCCESS: successfully create payment method",
			input: &paymentMethodModel.CreatePaymentMethodRequest{
				UUID: validUUID,
			},
			mocksSetup: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository) {
				paymentMethodRepo.On(
					"CreatePaymentMethod",
					mock.Anything,
					mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest"),
				).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "ERROR: repository returns error",
			input: &paymentMethodModel.CreatePaymentMethodRequest{
				UUID: validUUID,
			},
			mocksSetup: func(paymentMethodRepo *repositoryMocks.IPaymentMethodRepository) {
				paymentMethodRepo.On(
					"CreatePaymentMethod",
					mock.Anything,
					mock.AnythingOfType("*paymentMethodModel.CreatePaymentMethodRequest"),
				).Return(c.ErrSomeErrorForUnitTest)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			paymentMethodRepo := repositoryMocks.NewIPaymentMethodRepository(t)
			snapCoreRepo := repositoryMocks.NewISnapCoreRepository(t)
			creditCardRepo := repositoryMocks.NewICreditcardCoreProcessorRepository(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			tc.mocksSetup(paymentMethodRepo)

			svc := New(mockLogger, paymentMethodRepo, snapCoreRepo, creditCardRepo)

			ctx := context.Background()
			err := svc.Create(ctx, tc.input)

			if tc.wantErr {
				assert.Error(t, err)
				code, _ := pkgErr.ExtractError(err)
				assert.Equal(t, response.HttpErrDatabase, code)
			} else {
				assert.NoError(t, err)
				// Verify that UUID was set (should not be empty after Create)
				assert.NotEmpty(t, tc.input.UUID)
			}

			paymentMethodRepo.AssertExpectations(t)
		})
	}
}
