package merchantRcn

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Run("SUCCESS: Create repository without options", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		repo := New(mockMysql, mockLogger)

		assert.NotNil(t, repo)
	})

	t.Run("SUCCESS: Create repository with options", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		testConfig := &config.Config{
			ServiceName: "test-service",
		}

		// Create an option function that sets the config
		withConfig := func(r *MerchantRcnRepository) {
			r.config = testConfig
		}

		repo := New(mockMysql, mockLogger, withConfig).(*MerchantRcnRepository)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.config)
		assert.Equal(t, testConfig.ServiceName, repo.config.ServiceName)
	})

	t.Run("SUCCESS: Create repository with multiple options", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		testConfig := &config.Config{
			ServiceName: "test-service",
			Environment: "test",
		}

		// Create multiple option functions
		withConfig := func(r *MerchantRcnRepository) {
			r.config = testConfig
		}

		withLogger := func(r *MerchantRcnRepository) {
			r.logger = mockLogger
		}

		repo := New(mockMysql, mockLogger, withConfig, withLogger).(*MerchantRcnRepository)

		assert.NotNil(t, repo)
		assert.NotNil(t, repo.config)
		assert.Equal(t, testConfig.ServiceName, repo.config.ServiceName)
		assert.Equal(t, testConfig.Environment, repo.config.Environment)
	})
}
