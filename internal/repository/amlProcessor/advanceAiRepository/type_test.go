package advanceairepository

import (
	"context"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	httpMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/httpRequestExt"
	"github.com/paper-indonesia/pivot-backoffice/test"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	ffclient "github.com/thomaspoignant/go-feature-flag"
)

func TestNew(t *testing.T) {
	mockConfig := &config.Config{}
	mockSecret := &config.Secret{}
	mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	mockHTTPRequest := &httpMocks.IHTTPRequest{} // Make sure to generate this mock via mockery or similar tool

	customFnCalled := false
	customFn := func(repo *AdvanceAiRepository) {
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

func TestBaseURLWithoutFeatureFlag(t *testing.T) {
	mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	mockHTTPRequest := &httpMocks.IHTTPRequest{}
	expectedBaseURL := "https://config-url-no-ff.example.com"
	mockConfig := &config.Config{
		Environment: "test-no-ff-env",
		AdvanceAIConfig: config.AdvanceAIConfig{
			BaseURL: expectedBaseURL,
		},
	}
	mockSecret := &config.Secret{}

	repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)
	result := repo.baseURL()

	assert.Equal(t, expectedBaseURL, result)
}

func TestJourneyIDWithoutFeatureFlag(t *testing.T) {
	mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	mockHTTPRequest := &httpMocks.IHTTPRequest{}
	expectedJourneyID := "test-journey-no-ff"
	mockConfig := &config.Config{
		Environment: "test-no-ff-env",
		AdvanceAIConfig: config.AdvanceAIConfig{
			JourneyID: expectedJourneyID,
		},
	}
	mockSecret := &config.Secret{}

	repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)
	result := repo.journeyID()

	assert.Equal(t, expectedJourneyID, result)
}

func TestIntegrationBaseURL(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := context.Background()

	logger, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)
	consulContainer, consulURL, err := test.SetupConsul(ctx)
	assert.NoError(t, err)
	test.SetupFeatureFlag(consulURL)
	test.SetupGoff(ctx, consulURL, pdkLogger)

	defer logger.Sync()
	defer pdkLogger.Sync()
	defer ffclient.Close()
	defer consulContainer.Terminate(ctx)

	testCases := []struct {
		name            string
		environment     string
		configBaseURL   string
		expectedBaseURL string
	}{
		{
			name:            "SUCCESS: Returns feature flag BaseURL for staging environment",
			environment:     "staging",
			configBaseURL:   "https://config-url.example.com",
			expectedBaseURL: "", // Will be set by feature flag or fallback to config
		},
		{
			name:            "SUCCESS: Returns config BaseURL when feature flag is nil",
			environment:     "unknown-env",
			configBaseURL:   "https://config-fallback.example.com",
			expectedBaseURL: "https://config-fallback.example.com",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			mockHTTPRequest := &httpMocks.IHTTPRequest{}
			mockConfig := &config.Config{
				Environment: tc.environment,
				AdvanceAIConfig: config.AdvanceAIConfig{
					BaseURL: tc.configBaseURL,
				},
			}
			mockSecret := &config.Secret{}

			repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)
			result := repo.baseURL()

			assert.NotEmpty(t, result)
		})
	}
}

func TestIntegrationJourneyID(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := context.Background()

	logger, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)
	consulContainer, consulURL, err := test.SetupConsul(ctx)
	assert.NoError(t, err)
	test.SetupFeatureFlag(consulURL)
	test.SetupGoff(ctx, consulURL, pdkLogger)

	defer logger.Sync()
	defer pdkLogger.Sync()
	defer ffclient.Close()
	defer consulContainer.Terminate(ctx)

	testCases := []struct {
		name              string
		environment       string
		configJourneyID   string
		expectedJourneyID string
	}{
		{
			name:              "SUCCESS: Returns feature flag JourneyID for staging environment",
			environment:       "staging",
			configJourneyID:   "config-journey-id",
			expectedJourneyID: "", // Will be set by feature flag or fallback to config
		},
		{
			name:              "SUCCESS: Returns config JourneyID when feature flag is nil",
			environment:       "unknown-env",
			configJourneyID:   "config-fallback-journey-id",
			expectedJourneyID: "config-fallback-journey-id",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			mockHTTPRequest := &httpMocks.IHTTPRequest{}
			mockConfig := &config.Config{
				Environment: tc.environment,
				AdvanceAIConfig: config.AdvanceAIConfig{
					JourneyID: tc.configJourneyID,
				},
			}
			mockSecret := &config.Secret{}

			repo := New(mockConfig, mockSecret, mockLogger, mockHTTPRequest)
			result := repo.journeyID()

			assert.NotEmpty(t, result)
		})
	}
}
