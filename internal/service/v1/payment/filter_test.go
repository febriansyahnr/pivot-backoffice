package paymentService

import (
	"context"
	"testing"

	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFilterPaymentHistory(t *testing.T) {
	var (
		ctx             = context.Background()
		mockPaymentRepo = new(repositoryMocks.IPaymentRepository)
		paymentService  = PaymentService{
			paymentRepo: mockPaymentRepo,
		}
	)

	testCases := []struct {
		name      string
		payload   paymentModel.FilterPaymentHistoryOption
		callMock  func(t *testing.T)
		shouldErr bool
	}{
		{
			name: "SUCCESS: User have the payment histories",
			callMock: func(t *testing.T) {
				mockPaymentRepo.Mock.Test(t)
				mockPaymentRepo.On("FilterPaymentHistory", mock.Anything, mock.Anything).
					Return(&commonModel.PaginationResponse{}, nil)
			},
		},
		{
			name: "ERROR: Failed validation",
			payload: paymentModel.FilterPaymentHistoryOption{
				PaymentMethod: "STONES",
			},
			callMock: func(t *testing.T) {
			},
			shouldErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.callMock(t)
			_, err := paymentService.FilterPaymentHistory(ctx, tc.payload)

			if tc.shouldErr {
				assert.NotNil(t, err)
				assert.Error(t, err)
				return
			}

			assert.Nil(t, err)
		})
	}
}
