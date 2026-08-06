package payment

import (
	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	"github.com/paper-indonesia/pdk/go/monitoring"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/shopspring/decimal"
)

// createTestController creates a PaymentController instance for testing
func createTestController(paymentSvc service.IPaymentService, unifiedSvc service.IUnifiedPaymentService) *PaymentController {
	return &PaymentController{
		config: &config.Config{
			UnifiedPaymentConfig: config.UnifiedPaymentConfig{
				VirtualAccountConfig: &config.UnifiedPaymentVirtualAccountConfig{
					MinAmount: func() *float64 { v := 10000.0; return &v }(),
					MaxAmount: func() *float64 { v := 10000000.0; return &v }(),
				},
				EwalletConfig: &config.UnifiedPaymentEwalletConfig{
					MinAmount: func() *float64 { v := 5000.0; return &v }(),
					MaxAmount: func() *float64 { v := 5000000.0; return &v }(),
				},
				CardConfig: &config.UnifiedPaymentCardConfig{
					MinAmount: func() *float64 { v := 1000.0; return &v }(),
					MaxAmount: func() *float64 { v := 20000000.0; return &v }(),
				},
				QrConfig: &config.UnifiedPaymentQrConfig{
					MinAmount: func() *float64 { v := 1000.0; return &v }(),
					MaxAmount: func() *float64 { v := 5000000.0; return &v }(),
				},
			},
		},
		validate:              validator.New(),
		monitor:               &monitoring.Monitor{},
		paymentService:        paymentSvc,
		unifiedPaymentService: unifiedSvc,
		logger:                func() loggerMocks.ILogger { logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{}); return logger }(),
	}
}

// createTestPaymentResponse creates a sample PaymentResponse for testing
func createTestPaymentResponse(uuid, merchantID string, amount float64) *paymentModel.PaymentResponse {
	return &paymentModel.PaymentResponse{
		UUID:       uuid,
		MerchantID: merchantID,
		Status:     "PENDING",
		TotalAmount: &paymentModel.Amount{
			Value:    decimal.NewFromFloat(amount),
			Currency: "IDR",
		},
		PaymentMethod: "VIRTUAL_ACCOUNT",
	}
}

// createTestUnifiedResponse creates a sample UnifiedPaymentSessionResponse for testing
func createTestUnifiedResponse(id, refID string, amount float64) *unifiedPaymentModel.UnifiedPaymentSessionResponse {
	return &unifiedPaymentModel.UnifiedPaymentSessionResponse{
		ID:                id,
		ClientReferenceID: refID,
		Status:            "PROCESSING",
		Amount: unifiedPaymentModel.Amount{
			Value:    amount,
			Currency: "IDR",
		},
	}
}