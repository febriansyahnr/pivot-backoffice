package fds

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/paper-indonesia/pivot-backoffice/config"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	config := config.Config{}
	mockSvc := mocks.NewIFdsService(t)
	mockValidator := validator.New()
	mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

	customFnCalled := false
	customFn := func(c *InternalFdsController) {
		customFnCalled = true
	}

	controller := New(&config, mockLogger, mockValidator, mockSvc, customFn)

	assert.NotNil(t, controller)
	assert.True(t, customFnCalled, "customFn should have been called")
}
