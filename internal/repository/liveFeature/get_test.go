package liveFeature

import (
	"context"
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	mysqlMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/mySqlExt"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockConsulRetriever is a mock implementation of ConsulRetriever
type mockConsulRetriever struct {
	mock.Mock
}

func (m *mockConsulRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func TestNew(t *testing.T) {
	t.Run("SUCCESS: Create new repository", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		repo := New(mockMysql, mockLogger)

		assert.NotNil(t, repo)
	})
}

func TestWithConfig(t *testing.T) {
	t.Run("SUCCESS: Set config", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		repo := New(mockMysql, mockLogger).(*LiveFeatureRepository)

		cfg := &config.Config{
			FeatureFlagConfig: config.FeatureFlagConfig{
				ConsulAddr:       "http://localhost:8500",
				ConsulAppVersion: "app/version",
			},
		}

		repo.WithConfig(cfg)

		assert.Equal(t, cfg, repo.config)
	})
}

func TestWithSecret(t *testing.T) {
	t.Run("SUCCESS: Set secret", func(t *testing.T) {
		mockMysql := mysqlMocks.NewIMySqlExt(t)
		mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

		repo := New(mockMysql, mockLogger).(*LiveFeatureRepository)

		secret := &config.Secret{
			ConsulSecret: config.ConsulSecret{
				Token: "test-token",
			},
		}

		repo.WithSecret(secret)

		assert.Equal(t, secret, repo.secret)
	})
}

func TestDefaultConsulRetrieverFactory(t *testing.T) {
	t.Run("Test default consul retriever factory", func(t *testing.T) {
		// Call the default factory with dummy values
		// This tests that the function can be called and returns a retriever
		retriever, err := defaultConsulRetrieverFactory("http://localhost:8500", "key", "token")

		// The function should succeed in creating the retriever object
		// Even if connection to Consul fails later
		assert.NoError(t, err)
		assert.NotNil(t, retriever)
	})
}

func TestGetList(t *testing.T) {
	// merchantID := uuid.NewString()
	testCase := []struct {
		name      string
		mockSetup func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Get List without any filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List without any filter and total items is zero",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: Get List with merchantID filter",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)
			},
			wantErr: false,
		},
		{
			name: "FAILED: Get List on error get table",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(nil, errors.New("invalid table name"))

			},
			wantErr: true,
		},
		{
			name: "Get List on error retrieving data",
			mockSetup: func(mysqlMock *mysqlMocks.IMySqlExt, mockMySqlRows *mysqlMocks.IMySqlRows) {
				mockMySqlRows.On("Next").Return(true).Times(1)
				mockMySqlRows.On("Next").Return(false)
				mockMySqlRows.On("Close").Return(nil)
				mockMySqlRows.On("Scan",
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(errors.New("invalid scan data"))

				mysqlMock.On(
					"QueryContext",
					mock.AnythingOfType(constant.MockTypeValueContextReference),
					mock.AnythingOfType("string"),
				).Return(mockMySqlRows, nil)

			},
			wantErr: true,
		},
	}

	for _, tc := range testCase {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockMysqlRows := mysqlMocks.NewIMySqlRows(t)
			tc.mockSetup(mockMysql, mockMysqlRows)

			repo := New(mockMysql, mockLogger)
			ctx := context.Background()
			_, err := repo.GetAll(ctx)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockMysql.AssertExpectations(t)

		})
	}
}

func TestRetrieveAppVersion(t *testing.T) {
	testCases := []struct {
		name      string
		setupMock func(repo *LiveFeatureRepository)
		wantErr   bool
	}{
		{
			name: "SUCCESS: Retrieve app version",
			setupMock: func(repo *LiveFeatureRepository) {
				repo.consulRetrieverFactory = func(consulAddr, key, token string) (ConsulRetriever, error) {
					mockRetriever := new(mockConsulRetriever)
					validJSON := []byte(`{"versions":{"ios":"1.0.0","android":"1.0.1"}}`)
					mockRetriever.On("Retrieve", mock.Anything).Return(validJSON, nil)
					return mockRetriever, nil
				}
			},
			wantErr: false,
		},
		{
			name: "ERROR: NewConsulRetriever fails",
			setupMock: func(repo *LiveFeatureRepository) {
				repo.consulRetrieverFactory = func(consulAddr, key, token string) (ConsulRetriever, error) {
					return nil, errors.New("consul connection failed")
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Retrieve fails",
			setupMock: func(repo *LiveFeatureRepository) {
				repo.consulRetrieverFactory = func(consulAddr, key, token string) (ConsulRetriever, error) {
					mockRetriever := new(mockConsulRetriever)
					mockRetriever.On("Retrieve", mock.Anything).Return(nil, errors.New("retrieve failed"))
					return mockRetriever, nil
				}
			},
			wantErr: true,
		},
		{
			name: "ERROR: Invalid JSON response",
			setupMock: func(repo *LiveFeatureRepository) {
				repo.consulRetrieverFactory = func(consulAddr, key, token string) (ConsulRetriever, error) {
					mockRetriever := new(mockConsulRetriever)
					invalidJSON := []byte(`{invalid json}`)
					mockRetriever.On("Retrieve", mock.Anything).Return(invalidJSON, nil)
					return mockRetriever, nil
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockMysql := mysqlMocks.NewIMySqlExt(t)
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})

			repo := New(mockMysql, mockLogger).(*LiveFeatureRepository)

			// Setup config and secret
			repo.WithConfig(&config.Config{
				FeatureFlagConfig: config.FeatureFlagConfig{
					ConsulAddr:       "http://localhost:8500",
					ConsulAppVersion: "app/version",
				},
			})
			repo.WithSecret(&config.Secret{
				ConsulSecret: config.ConsulSecret{
					Token: "test-token",
				},
			})

			tc.setupMock(repo)

			ctx := context.Background()
			result, err := repo.RetrieveAppVersion(ctx)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
