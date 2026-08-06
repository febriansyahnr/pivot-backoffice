package fraudruleservice

import (
	"testing"

	mockRepositories "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	fraudRulesRepo := mockRepositories.NewIFraudRulesRepository(t)
	logger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	customFnCalled := false
	customFn := func(repo *FraudRuleService) {
		customFnCalled = true
	}

	service := New(
		logger,
		fraudRulesRepo,
		customFn,
	)

	assert.NotNil(t, service)
	assert.True(t, customFnCalled, "customFn should have been called")
}
