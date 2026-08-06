package paymentService

import (
	"context"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestChangePaymentMethod(t *testing.T) {
	internalFunc := serviceMocks.NewIPaymentInternalDirectFunc(t)

	service := &PaymentService{internal: internalFunc}

	request := &paymentModel.UpdateUnifiedPaymentRequest{
		Amount: &paymentModel.Amount{}, ExpiryAt: &time.Time{},
	}

	tests := []struct {
		name       string
		setupMock  func()
		wantErr    error
		wantResult *paymentModel.UpdateUnifiedPaymentResponse
	}{
		{
			name: "ERROR:Some error",
			setupMock: func() {
				internalFunc.On(
					"CreateUnifiedPayment", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantErr: c.ErrSomeErrorForUnitTest,
		},
		{
			name: "SUCCESS",
			setupMock: func() {
				internalFunc.On(
					"CreateUnifiedPayment", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.CreateUnifiedPaymentResponse{
					ID:                "3f4871c1-e9dc-418a-9512-79ddc30af281",
					ClientReferenceID: "123456",
					Amount: paymentModel.Amount{
						Value:    decimal.NewFromFloat(1_250),
						Currency: "IDR",
					},
					PaymentMethod: "VIRTUAL_ACCOUNT",
					PaymentLink:   "https://",
				}, nil)
			},
			wantResult: &paymentModel.UpdateUnifiedPaymentResponse{
				ID:                "3f4871c1-e9dc-418a-9512-79ddc30af281",
				ClientReferenceID: "123456",
				Amount: paymentModel.Amount{
					Value:    decimal.NewFromFloat(1_250),
					Currency: "IDR",
				},
				PaymentMethod: "VIRTUAL_ACCOUNT",
				PaymentLink:   "https://",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.changePaymentMethod(context.Background(), request, &paymentModel.Payment{})
			assert.Equal(t, test.wantErr, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
