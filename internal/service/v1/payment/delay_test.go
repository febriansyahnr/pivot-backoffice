package paymentService

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentConstant "github.com/paper-indonesia/pivot-backoffice/constant/payment"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestPaymentService_getDelayedConfigDuration(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	service := &PaymentService{
		logger: logger,
		config: &config.Config{
			UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				ExpiringProcessedBackoffMinutes: []int{
					1, 3, 5, 10, 15, 30,
				},
			},
		},
	}

	testCases := []struct {
		name               string
		ctx                context.Context
		payment            *paymentModel.Payment
		expectedRetryCount int
		expectedDuration   time.Duration
	}{
		{
			name: "success - credit card payment with retry count 0",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(0)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   1 * time.Minute,
		},
		{
			name: "success - ewallet payment with retry count 1",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(1)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			expectedRetryCount: 1,
			expectedDuration:   3 * time.Minute,
		},
		{
			name: "success - credit card payment with retry count 2",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(2)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 2,
			expectedDuration:   5 * time.Minute,
		},
		{
			name: "success - ewallet payment with retry count 3",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(3)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			expectedRetryCount: 3,
			expectedDuration:   10 * time.Minute,
		},
		{
			name: "success - credit card payment with retry count 4",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(4)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 4,
			expectedDuration:   15 * time.Minute,
		},
		{
			name: "success - ewallet payment with retry count 5",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(5)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			expectedRetryCount: 5,
			expectedDuration:   30 * time.Minute,
		},
		{
			name:               "payment is nil",
			ctx:                context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(0)),
			payment:            nil,
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "payment method is virtual account - not eligible for delay",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(0)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_VIRTUAL_ACCOUNT,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "payment method is qris - not eligible for delay",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(0)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_QRIS,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "payment method is bank transfer - not eligible for delay",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(0)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_BANK_TRANSFER,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "context does not have retry count",
			ctx:  context.Background(),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "context has invalid retry count type",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, "invalid"),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
		{
			name: "retry count exceeds available configuration",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(10)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 10,
			expectedDuration:   -1,
		},
		{
			name: "retry count equals configuration length (edge case)",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(6)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_EWALLET,
				},
			},
			expectedRetryCount: 6,
			expectedDuration:   -1,
		},
		{
			name: "negative retry count",
			ctx:  context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(-1)),
			payment: &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			},
			expectedRetryCount: 0,
			expectedDuration:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualRetryCount, actualDuration := service.getDelayedConfigDuration(tc.ctx, tc.payment)

			assert.Equal(t, tc.expectedRetryCount, actualRetryCount, "retry count should match expected value")
			assert.Equal(t, tc.expectedDuration, actualDuration, "duration should match expected value")
		})
	}
}

func TestPaymentService_getDelayedConfigDuration_ConfigurationVerification(t *testing.T) {
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	service := &PaymentService{
		logger: logger,
		config: &config.Config{
			UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				ExpiringProcessedBackoffMinutes: []int{
					1, 3, 5, 10, 15, 30,
				},
			},
		},
	}

	// Test all configuration values to ensure they match expected durations
	expectedConfigurations := map[int]time.Duration{
		0: 1 * time.Minute,
		1: 3 * time.Minute,
		2: 5 * time.Minute,
		3: 10 * time.Minute,
		4: 15 * time.Minute,
		5: 30 * time.Minute,
	}

	for retryCount, expectedDuration := range expectedConfigurations {
		t.Run(fmt.Sprintf("verify_config_retry_%d", retryCount), func(t *testing.T) {
			ctx := context.WithValue(context.Background(), constant.CtxRabbitMQRetryCount, int32(retryCount))
			payment := &paymentModel.Payment{
				PaymentMethod: paymentModel.PaymentMethod{
					Type: paymentConstant.PAYMENT_METHOD_CREDIT_CARD,
				},
			}

			actualRetryCount, actualDuration := service.getDelayedConfigDuration(ctx, payment)

			assert.Equal(t, retryCount, actualRetryCount, "retry count should match")
			assert.Equal(t, expectedDuration, actualDuration, "duration should match configured value")
		})
	}
}
