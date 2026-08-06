package payment

import (
	"context"
	"errors"
	"testing"

	loggerMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/logger"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/stretchr/testify/mock"
)

func TestExpirePendingPayment(t *testing.T) {
	testCases := []struct {
		name      string
		mockSetup func(logger *loggerMocks.ILogger, paymentService *serviceMocks.IPaymentService)
	}{
		{
			name: "when succeeded to publish expiration payment, then should log info",
			mockSetup: func(logger *loggerMocks.ILogger, paymentService *serviceMocks.IPaymentService) {
				paymentService.On("PublishPaymentExpirationMessage", mock.Anything).Return(nil).Once()
				logger.On("Info", mock.Anything, "Successfully published payment expiration message").Once()
			},
		},
		{
			name: "when failed to publish expiration payment, then should log error",
			mockSetup: func(logger *loggerMocks.ILogger, paymentService *serviceMocks.IPaymentService) {
				paymentService.On("PublishPaymentExpirationMessage", mock.Anything).Return(errors.New("database error")).Once()
				logger.On("Fatal", mock.Anything, "Failed to publish payment expiration message", mock.Anything).Once()
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger := loggerMocks.NewILogger(t)
			mockPaymentService := serviceMocks.NewIPaymentService(t)

			tc.mockSetup(mockLogger, mockPaymentService)

			h := New(mockLogger, mockPaymentService)
			h.PublishPendingPaymentExpirationEvent(context.Background())
		})
	}
}
