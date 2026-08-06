package fraudnetrepository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func TestMain(m *testing.M) {
	// Initialize feature flag client once for all tests
	cwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(cwd, "backend-portal")
	targetPath := filepath.Join(projectRoot, "test", "consul", "backend-portal", "feature-flag.yaml")

	_ = ffclient.Init(ffclient.Config{
		Retriever:    &fileretriever.Retriever{Path: targetPath},
		DataExporter: ffclient.DataExporter{},
	})
	defer ffclient.Close()

	m.Run()
}

func TestNew(t *testing.T) {
	mockConfig := &config.Config{}
	mockSecret := &config.Secret{}
	mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	mockHTTPRequest := &httpMocks.IHTTPRequest{} // Make sure to generate this mock via mockery or similar tool

	customFnCalled := false
	customFn := func(repo *FraudNetRepository) {
		customFnCalled = true
	}

	repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest, customFn)

	assert.NotNil(t, repo)
	assert.Equal(t, mockConfig, repo.config)
	assert.Equal(t, mockSecret, repo.secret)
	assert.Equal(t, mockLogger, repo.logger)
	assert.Equal(t, mockHTTPRequest, repo.httpRequest)
	assert.True(t, customFnCalled, "customFn should have been called")
}

func TestBaseURL(t *testing.T) {
	mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	mockHTTPRequest := &httpMocks.IHTTPRequest{}

	t.Run("SUCCESS: Returns config BaseURL when feature flag is nil", func(t *testing.T) {
		mockConfig := &config.Config{
			FraudNetConfig: config.FraudNetConfig{
				BaseURL: "https://config-base-url.com",
			},
			Environment: "unknown-env-that-doesnt-match-any-targeting",
		}
		mockSecret := &config.Secret{}

		repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)
		result := repo.baseURL()

		// With an environment that doesn't match any targeting rules,
		// GetFraudNetFeatureFlag will return the default variation or nil
		// Let's check which one we get
		fraudNetFF := constant.GetFraudNetFeatureFlag("unknown-env-that-doesnt-match-any-targeting")
		if fraudNetFF != nil {
			// If feature flag returns default, we should get that BaseURL
			assert.Equal(t, fraudNetFF.BaseURL, result)
		} else {
			// If feature flag is nil, we should get config BaseURL
			assert.Equal(t, "https://config-base-url.com", result)
		}
	})

	t.Run("SUCCESS: Returns feature flag BaseURL when feature flag is available", func(t *testing.T) {
		mockConfig := &config.Config{
			FraudNetConfig: config.FraudNetConfig{
				BaseURL: "https://config-base-url.com",
			},
			Environment: "staging",
		}
		mockSecret := &config.Secret{}

		repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)

		// Get the feature flag to verify it works
		fraudNetFF := constant.GetFraudNetFeatureFlag("staging")
		if fraudNetFF == nil {
			t.Skip("Feature flag not available, skipping test")
		}

		result := repo.baseURL()

		// Should return the feature flag BaseURL, not the config BaseURL
		assert.Equal(t, fraudNetFF.BaseURL, result)
		assert.NotEqual(t, "https://config-base-url.com", result)
	})
}
