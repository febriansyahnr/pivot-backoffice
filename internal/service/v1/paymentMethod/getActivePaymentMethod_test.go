package paymentMethodService_test

import (
	"context"
	"testing"

	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/paymentMethod"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetActivePaymentMethodDetailForPaymentRequest(t *testing.T) {
	repo := repoMocks.NewIPaymentMethodRepository(t)

	service := New(nil, repo, nil, nil)

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *paymentModel.PaymentMethodWithPivot
	}{
		{
			name: "ERROR:Some error", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetActivePaymentMethodByRequest", mock.Anything, mock.Anything,
				).Once().Return(nil, assert.AnError)
			},
			wantErr: assert.AnError,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				repo.On(
					"GetActivePaymentMethodByRequest", mock.Anything, mock.Anything,
				).Once().Return(&paymentModel.PaymentMethodWithPivot{}, nil)
			},
			wantResult: &paymentModel.PaymentMethodWithPivot{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetActivePaymentMethodDetailForPaymentRequest(context.Background(), paymentModel.GetPaymentMethodFilterRequest{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
