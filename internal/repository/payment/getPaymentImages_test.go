package paymentRepository_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/repository/payment"
	"github.com/paper-indonesia/pivot-backoffice/test"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

// MockConsulRetrieverForImages is a mock implementation of ConsulRetriever for testing
type MockConsulRetrieverForImages struct {
	RetrieveFunc func(ctx context.Context) ([]byte, error)
}

func (m *MockConsulRetrieverForImages) Retrieve(ctx context.Context) ([]byte, error) {
	if m.RetrieveFunc != nil {
		return m.RetrieveFunc(ctx)
	}
	return nil, errors.New("mock not configured")
}

func TestIntegrationRetrieveImages(t *testing.T) {
	if os.Getenv(constant.IntegrationTestEnv) != "1" {
		t.Skip(constant.SkipIntegrationTest)
	}

	ctx := context.Background()
	logger, pdkLogger, err := test.SetupLogger()
	assert.NoError(t, err)
	defer logger.Sync()

	consulContainer, consulURL, err := test.SetupConsul(ctx)
	assert.NoError(t, err)
	defer func() { _ = consulContainer.Terminate(ctx) }()

	_ = test.SetupFeatureFlag(consulURL)
	_ = test.SetupGoff(ctx, consulURL, pdkLogger)
	_ = test.SetupConsulValue(consulURL, "payment-images")

	testCases := []struct {
		name    string
		wantErr bool
		setup   func(*config.Config, *config.Secret)
	}{
		{
			name:    "SUCCESS",
			wantErr: false,
			setup: func(cfg *config.Config, secret *config.Secret) {
				cfg.FeatureFlagConfig.ConsulAddr = consulURL
				cfg.FeatureFlagConfig.ConsulPaymentImages = "backend-portal/payment-images"
				secret.ConsulSecret.Token = ""
			},
		},
		{
			name:    "ERROR: Invalid consul address",
			wantErr: true,
			setup: func(cfg *config.Config, secret *config.Secret) {
				cfg.FeatureFlagConfig.ConsulAddr = "://invalid-url"
				cfg.FeatureFlagConfig.ConsulPaymentImages = "backend-portal/payment-images"
				secret.ConsulSecret.Token = ""
			},
		},
	}

	// Test with mock retriever that returns error during Retrieve
	t.Run("ERROR: Retriever returns error", func(t *testing.T) {
		mockRetriever := &MockConsulRetrieverForImages{
			RetrieveFunc: func(ctx context.Context) ([]byte, error) {
				return nil, errors.New("failed to retrieve from consul")
			},
		}

		cfg := &config.Config{
			FeatureFlagConfig: config.FeatureFlagConfig{
				ConsulAddr:          consulURL,
				ConsulPaymentImages: "backend-portal/payment-images",
			},
		}
		secret := &config.Secret{
			ConsulSecret: config.ConsulSecret{
				Token: "",
			},
		}

		repo := New(nil, pdkLogger, WithConsulRetriever(mockRetriever))
		repo.WithConfig(cfg)
		repo.WithSecret(secret)

		_, err := repo.RetrieveImages(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve from consul")
	})

	// Set up invalid JSON test separately to avoid affecting other tests
	t.Run("ERROR: Invalid JSON data in consul", func(t *testing.T) {
		// Set up consul client and add invalid JSON data
		consulConfig := api.DefaultConfig()
		consulConfig.Address = consulURL
		consulClient, err := api.NewClient(consulConfig)
		assert.NoError(t, err)

		// Add invalid JSON to Consul
		kv := consulClient.KV()
		p := &api.KVPair{Key: "backend-portal/payment-images-invalid", Value: []byte("invalid-json-data")}
		_, err = kv.Put(p, nil)
		assert.NoError(t, err)

		cfg := &config.Config{
			FeatureFlagConfig: config.FeatureFlagConfig{
				ConsulAddr:          consulURL,
				ConsulPaymentImages: "backend-portal/payment-images-invalid",
			},
		}
		secret := &config.Secret{
			ConsulSecret: config.ConsulSecret{
				Token: "",
			},
		}

		repo := New(nil, pdkLogger)
		repo.WithConfig(cfg)
		repo.WithSecret(secret)

		_, err = repo.RetrieveImages(context.Background())
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{}
			secret := &config.Secret{}

			// Call the setup function to configure test-specific settings
			tc.setup(cfg, secret)

			repo := New(nil, pdkLogger)
			repo.WithConfig(cfg)
			repo.WithSecret(secret)

			ctx := context.Background()
			_, err := repo.RetrieveImages(ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
