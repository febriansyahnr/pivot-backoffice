package unifiedPaymentService_test

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v2/unifiedPayment"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	loggerMocks "github.com/paper-indonesia/pdk/v2/logger"
)

func TestEncryptCard(t *testing.T) {
	testCases := []struct {
		name                  string
		request               *unifiedPaymentModel.EncryptCardRequest
		wantError             bool
		featureFlagEnabled    bool
		setupMock             func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository)
		expectedErrorContains string
	}{
		{
			name: "SUCCESS: encrypt card with feature flag enabled",
			request: &unifiedPaymentModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
				CardRequest: unifiedPaymentModel.EncryptCardDetailRequest{
					Number:      "1234567890123456",
					ExpiryMonth: "12",
					ExpiryYear:  "25",
					CVC:         "123",
					NameOnCard:  "Test User",
				},
				DeviceInformation: unifiedPaymentModel.DeviceInformation{
					Type:      "web",
					UserAgent: "test-agent",
					IpAddress: "127.0.0.1",
				},
			},
			wantError:          false,
			featureFlagEnabled: true,
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				mockCreditcardProcessorRepo.On(
					"EncryptCardData",
					mock.Anything,
					mock.Anything,
				).Return(&creditcardCoreProcessorModel.EncryptedCardResponse{
					ClientReferenceID: "test-ref-123",
					EncryptedCard:     "encrypted-card-data",
					EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
						First8Digits:     "12345678",
						First6Digits:     "123456",
						Last4Digits:      "7890",
						ExpiryMonth:      "12",
						ExpiryYear:       "25",
						HasAssociatedCVC: true,
						Fingerprint:      "test-fingerprint",
					},
					CreatedAt: "2023-01-01T00:00:00Z",
				}, nil)
			},
		},
		{
			name: "ERROR: feature flag disabled",
			request: &unifiedPaymentModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError:             true,
			featureFlagEnabled:    false,
			expectedErrorContains: "forbidden",
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				// No mock setup needed as feature flag check happens first
			},
		},
		{
			name: "ERROR: creditcard service error",
			request: &unifiedPaymentModel.EncryptCardRequest{
				MerchantID:        "test-merchant-id",
				ClientReferenceID: "test-ref-123",
			},
			wantError:          true,
			featureFlagEnabled: true,
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				mockCreditcardProcessorRepo.On(
					"EncryptCardData",
					mock.Anything,
					mock.Anything,
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCreditcardProcessorRepo := repoMocks.NewICreditcardCoreProcessorRepository(t)
			tc.setupMock(mockCreditcardProcessorRepo)

			config := &config.Config{
				Environment: "test",
			}

			svc := New(config, mockLogger, nil, nil, nil, WithCreditCardCoreProcessorRepo(mockCreditcardProcessorRepo))

			ffContentConfig := `
backend-portal-card-encryption-whitelisted-merchant:
  variations:
    ON: true
    OFF: false
  targeting:
    - name: allowed environemnt
      query: environment in ["local", "staging"]
      variation: ON
    - name: Check whitelisted merchant id
      query: merchant_id in ["test-merchant-id"]
      variation: ON
  defaultRule:
    variation: OFF`
			f, err := os.CreateTemp(os.TempDir(), "encryption-card-merchant-*.yaml")
			require.NoError(t, err)
			defer func() { require.NoError(t, os.Remove(f.Name())) }()
			defer func() { require.NoError(t, f.Close()) }()

			_, err = f.WriteString(ffContentConfig)
			require.NoError(t, err)

			err = ffclient.Init(ffclient.Config{
				FileFormat: "YAML",
				Retriever: &fileretriever.Retriever{
					Path: f.Name(),
				},
			})
			require.NoError(t, err)
			defer ffclient.Close()

			// For feature flag disabled tests, we skip the actual service call
			// since feature flag testing should be done separately
			if !tc.featureFlagEnabled {
				tc.request.MerchantID = "test-merchant-id-fail"
			}

			result, err := svc.EncryptCard(context.Background(), tc.request)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test-ref-123", result.ClientReferenceID)
				assert.Equal(t, "encrypted-card-data", result.EncryptedCard)
			}

			if tc.featureFlagEnabled || tc.expectedErrorContains != "forbidden" {
				mockCreditcardProcessorRepo.AssertExpectations(t)
			}
		})
	}
}

func TestGetEncryptedCard(t *testing.T) {
	testCases := []struct {
		name                  string
		merchantId            string
		cardId                string
		wantError             bool
		featureFlagEnabled    bool
		setupMock             func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository)
		expectedErrorContains string
	}{
		{
			name:               "SUCCESS: get encrypted card with feature flag enabled",
			merchantId:         "test-merchant-id",
			cardId:             "test-card-id",
			wantError:          false,
			featureFlagEnabled: true,
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				mockCreditcardProcessorRepo.On(
					"GetEncryptedCardData",
					mock.Anything,
					"test-merchant-id",
					"test-card-id",
				).Return(&creditcardCoreProcessorModel.EncryptedCardResponse{
					ClientReferenceID: "test-ref-123",
					EncryptedCard:     "encrypted-card-data",
					EncryptedCardInformation: creditcardCoreProcessorModel.EncryptedCardInformationResponse{
						First8Digits:     "12345678",
						First6Digits:     "123456",
						Last4Digits:      "7890",
						ExpiryMonth:      "12",
						ExpiryYear:       "25",
						HasAssociatedCVC: true,
						Fingerprint:      "test-fingerprint",
					},
					CreatedAt: "2023-01-01T00:00:00Z",
				}, nil)
			},
		},
		{
			name:                  "ERROR: feature flag disabled",
			merchantId:            "test-merchant-id",
			cardId:                "test-card-id",
			wantError:             true,
			featureFlagEnabled:    false,
			expectedErrorContains: "forbidden",
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				// No mock setup needed as feature flag check happens first
			},
		},
		{
			name:               "ERROR: creditcard service error",
			merchantId:         "test-merchant-id",
			cardId:             "test-card-id",
			wantError:          true,
			featureFlagEnabled: true,
			setupMock: func(mockCreditcardProcessorRepo *repoMocks.ICreditcardCoreProcessorRepository) {
				mockCreditcardProcessorRepo.On(
					"GetEncryptedCardData",
					mock.Anything,
					"test-merchant-id",
					"test-card-id",
				).Return(nil, constant.ErrSomeErrorForUnitTest)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockLogger, _ := loggerMocks.NewZapLogger(loggerMocks.Config{})
			mockCreditcardProcessorRepo := repoMocks.NewICreditcardCoreProcessorRepository(t)
			tc.setupMock(mockCreditcardProcessorRepo)

			config := &config.Config{
				Environment: "test",
			}

			svc := New(config, mockLogger, nil, nil, nil, WithCreditCardCoreProcessorRepo(mockCreditcardProcessorRepo))

			ffContentConfig := `
backend-portal-card-encryption-whitelisted-merchant:
  variations:
    ON: true
    OFF: false
  targeting:
    - name: allowed environemnt
      query: environment in ["local", "staging"]
      variation: ON
    - name: Check whitelisted merchant id
      query: merchant_id in ["test-merchant-id"]
      variation: ON
  defaultRule:
    variation: OFF`
			f, err := os.CreateTemp(os.TempDir(), "encryption-card-merchant-*.yaml")
			require.NoError(t, err)
			defer func() { require.NoError(t, os.Remove(f.Name())) }()
			defer func() { require.NoError(t, f.Close()) }()

			_, err = f.WriteString(ffContentConfig)
			require.NoError(t, err)

			err = ffclient.Init(ffclient.Config{
				FileFormat: "YAML",
				Retriever: &fileretriever.Retriever{
					Path: f.Name(),
				},
			})
			require.NoError(t, err)
			defer ffclient.Close()

			// For feature flag disabled tests, we skip the actual service call
			// since feature flag testing should be done separately
			if !tc.featureFlagEnabled {
				tc.merchantId = "test-merchant-id-fail"
			}

			result, err := svc.GetEncryptedCard(context.Background(), tc.merchantId, tc.cardId)
			WithCreditCardCoreProcessorRepo(mockCreditcardProcessorRepo)

			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "test-ref-123", result.ClientReferenceID)
				assert.Equal(t, "encrypted-card-data", result.EncryptedCard)
			}

			if tc.featureFlagEnabled || tc.expectedErrorContains != "forbidden" {
				mockCreditcardProcessorRepo.AssertExpectations(t)
			}
		})
	}
}
