package fdsservice

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/repository"
	rabbitMqMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestNew(t *testing.T) {
	accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
	paymentRepo := mockRepositories.NewIPaymentRepository(t)
	merchantRepo := mockRepositories.NewIMerchantRepository(t)
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
	fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
	processors := map[string]repository.IFdsProcessorRepository{
		constant.PROVIDER_FRAUD_NET: fraudNetMock,
	}
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	customFnCalled := false
	customFn := func(repo *FdsService) {
		customFnCalled = true
	}
	service := New(
		&config.Config{
			FdsConfig: config.FdsConfig{
				ScoreThreshold: 50,
			},
		},
		logger,
		fraudRulesRepo,
		ruleEvalRepo,
		accountTrxRepo,
		paymentRepo,
		merchantRepo,
		processors,
		customFn,
	)
	assert.NotNil(t, service)
	assert.True(t, customFnCalled, "customFn should have been called")
}

func TestWithRabbitMqExt(t *testing.T) {
	// Setup
	accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
	paymentRepo := mockRepositories.NewIPaymentRepository(t)
	merchantRepo := mockRepositories.NewIMerchantRepository(t)
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
	fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
	processors := map[string]repository.IFdsProcessorRepository{
		constant.PROVIDER_FRAUD_NET: fraudNetMock,
	}
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
	rabbitMq := rabbitMqMocks.NewRabbitMQExt(t)

	// Create service with RabbitMQ dependency
	service := New(
		&config.Config{},
		logger,
		fraudRulesRepo,
		ruleEvalRepo,
		accountTrxRepo,
		paymentRepo,
		merchantRepo,
		processors,
		WithRabbitMqExt(rabbitMq),
	)

	// Assert
	fdsService, ok := service.(*FdsService)
	assert.True(t, ok, "service should be of type *FdsService")
	assert.NotNil(t, fdsService.rabbitMq, "rabbitMq should be set")
	assert.Equal(t, rabbitMq, fdsService.rabbitMq, "rabbitMq should be equal to the one provided")
}

func TestWithOutboundRepository(t *testing.T) {
	// Setup
	accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
	paymentRepo := mockRepositories.NewIPaymentRepository(t)
	merchantRepo := mockRepositories.NewIMerchantRepository(t)
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
	fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
	outboundRepo := mockRepositories.NewIOutboundRepository(t)
	processors := map[string]repository.IFdsProcessorRepository{
		constant.PROVIDER_FRAUD_NET: fraudNetMock,
	}
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// Create service with OutboundRepository dependency
	service := New(
		&config.Config{},
		logger,
		fraudRulesRepo,
		ruleEvalRepo,
		accountTrxRepo,
		paymentRepo,
		merchantRepo,
		processors,
		WithOutboundRepository(outboundRepo),
	)

	// Assert
	fdsService, ok := service.(*FdsService)
	assert.True(t, ok, "service should be of type *FdsService")
	assert.NotNil(t, fdsService.outboundRepository, "outboundRepository should be set")
	assert.Equal(t, outboundRepo, fdsService.outboundRepository, "outboundRepository should be equal to the one provided")
}

func TestMultipleDependencies(t *testing.T) {
	// Setup
	accountTrxRepo := mockRepositories.NewIAccountTransactionRepository(t)
	paymentRepo := mockRepositories.NewIPaymentRepository(t)
	merchantRepo := mockRepositories.NewIMerchantRepository(t)
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	ruleEvalRepo := mockRepositories.NewIRuleEvaluationsRepository(t)
	fraudNetMock := mockRepositories.NewIFdsProcessorRepository(t)
	outboundRepo := mockRepositories.NewIOutboundRepository(t)
	rabbitMq := rabbitMqMocks.NewRabbitMQExt(t)
	processors := map[string]repository.IFdsProcessorRepository{
		constant.PROVIDER_FRAUD_NET: fraudNetMock,
	}
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	// Create service with multiple dependencies
	service := New(
		&config.Config{},
		logger,
		fraudRulesRepo,
		ruleEvalRepo,
		accountTrxRepo,
		paymentRepo,
		merchantRepo,
		processors,
		WithOutboundRepository(outboundRepo),
		WithRabbitMqExt(rabbitMq),
	)

	// Assert
	fdsService, ok := service.(*FdsService)
	assert.True(t, ok, "service should be of type *FdsService")
	assert.NotNil(t, fdsService.outboundRepository, "outboundRepository should be set")
	assert.Equal(t, outboundRepo, fdsService.outboundRepository, "outboundRepository should be equal to the one provided")
	assert.NotNil(t, fdsService.rabbitMq, "rabbitMq should be set")
	assert.Equal(t, rabbitMq, fdsService.rabbitMq, "rabbitMq should be equal to the one provided")
}

func TestPackageVariables(t *testing.T) {
	// Test that the tracer is initialized correctly
	expectedTracer := otel.Tracer("FdsService")
	assert.NotNil(t, tracer, "tracer should be initialized")
	// This is a limited check as we can't easily compare tracer instances
	assert.IsType(t, expectedTracer, tracer, "tracer should be of the expected type")
}
