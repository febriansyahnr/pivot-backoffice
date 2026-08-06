package paymentService_test

import (
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/payment"
	encryptMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/encrypt"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	cryptoProviderMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	redisExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	vaultMocks "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	"github.com/paper-indonesia/pivot-backoffice/pkg/encryption"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetEncryptionKey(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	cache := redisExtMock.NewIRedisExt(t)
	cryptoAesGcm := encryptMocks.NewEncrypter(t)
	secretManager := vaultMocks.NewIVaultKeyValue(t)

	service := New(nil, logger, nil, nil, nil, nil, nil, WithRedisClient(cache), WithCryptoAesGcm(cryptoAesGcm), WithSecretManager(secretManager))

	encryptionKey := "ZW5jcnlwdGlvbi1rZXk="

	tests := []struct {
		name       string
		setupMock  func()
		wantError  error
		wantResult *paymentModel.GetEncryptionKeyResponse
	}{
		{
			name: "ERROR: Get encryption key from cache", // NOSONAR
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetErr(assert.AnError)

				cache.On("Get", mock.Anything, constant.PaymentEncryptionKeyCacheKey).Once().Return(result)
				logger.On("Error", mock.Anything, "Failed to get encryption key from cache", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrDatabase, constant.ErrInternalServerForUser),
		},
		{
			name: "SUCCESS: Encryption key found", // NOSONAR
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetVal(encryptionKey)

				cache.On("Get", mock.Anything, constant.PaymentEncryptionKeyCacheKey).Once().Return(result)
			},
			wantResult: &paymentModel.GetEncryptionKeyResponse{EncryptionKey: encryptionKey},
		},
		{
			name: "ERROR: Get public key from secret manager", // NOSONAR
			setupMock: func() {
				result := &redis.StringCmd{}
				result.SetErr(redis.Nil)

				cache.On("Get", mock.Anything, constant.PaymentEncryptionKeyCacheKey).Return(result)
				secretManager.On("GetSecretKeyString", mock.Anything, "public_key").Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to get public key from secret manager", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR: Encrypt public key", // NOSONAR
			setupMock: func() {
				secretManager.On("GetSecretKeyString", mock.Anything, "public_key").Return(&vault.SecretKey[string]{Value: "public_key"}, nil)
				cryptoAesGcm.On("Encrypt", "public_key").Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to encrypt public key with crypto aes-gcm", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "WARN: Store encryption key to cache", // NOSONAR
			setupMock: func() {
				result := &redis.StatusCmd{}
				result.SetErr(assert.AnError)

				cryptoAesGcm.On("Encrypt", "public_key").Return([]byte("encryption-key"), nil)
				cache.On("Set", mock.Anything, constant.PaymentEncryptionKeyCacheKey, encryptionKey, constant.PaymentEncryptionKeyTTL).Once().Return(result)
				logger.On("Warn", mock.Anything, "Failed to store encryption key to cache", mock.Anything).Once().Return()
			},
			wantResult: &paymentModel.GetEncryptionKeyResponse{EncryptionKey: encryptionKey},
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				cache.On("Set", mock.Anything, constant.PaymentEncryptionKeyCacheKey, encryptionKey, constant.PaymentEncryptionKeyTTL).Once().Return(&redis.StatusCmd{})
			},
			wantResult: &paymentModel.GetEncryptionKeyResponse{EncryptionKey: encryptionKey},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			result, err := service.GetEncryptionKey(t.Context())
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)

			cache.AssertExpectations(t)
			logger.AssertExpectations(t)
			cryptoAesGcm.AssertExpectations(t)
			secretManager.AssertExpectations(t)
		})
	}
}

func TestDecryptRequest(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	secretManager := vaultMocks.NewIVaultKeyValue(t)
	cryptoProvider := cryptoProviderMock.NewCryptoProvider(t)

	service := New(
		nil, logger, nil, nil, nil, nil, nil,
		WithSecretManager(secretManager),
		WithCryptoProvider(cryptoProvider),
	)

	data := &encryption.DataEncryption{}

	tests := []struct {
		name       string
		dst        map[string]string
		setupMock  func()
		wantError  error
		wantResult map[string]string
	}{
		{
			name: "ERROR:Get private key from secret manager", // NOSONAR
			setupMock: func() {
				secretManager.On("GetSecretKeyString", mock.Anything, "private_key").Once().Return(nil, assert.AnError)
				logger.On("Error", mock.Anything, "Failed to get private key from secret manager", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrInternal, constant.ErrInternalServerForUser),
		},
		{
			name: "ERROR:Decrypt request", // NOSONAR
			setupMock: func() {
				secretManager.On(
					"GetSecretKeyString", mock.Anything, "private_key",
				).Return(&vault.SecretKey[string]{Value: "private-key-pem"}, nil)
				cryptoProvider.On("DecryptHybrid", data).Once().Return(nil, assert.AnError)
				logger.On("Warn", mock.Anything, "Failed to decrypt request payload", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload),
		},
		{
			name: "ERROR:Unmarshal decrypt result to destination", // NOSONAR
			setupMock: func() {
				cryptoProvider.On("DecryptHybrid", data).Once().Return([]byte(`ABC`), nil)
				logger.On("Warn", mock.Anything, "Failed to unmarshal plaintext to destination", mock.Anything).Once().Return()
			},
			wantError: pkgErrs.New(response.HttpErrRequest, constant.ErrMalformedRequestBodyPayload),
		},
		{
			name: "SUCCESS", // NOSONAR
			dst:  map[string]string{},
			setupMock: func() {
				cryptoProvider.On("DecryptHybrid", data).Once().Return([]byte(`{"message":"ok"}`), nil)
			},
			wantResult: map[string]string{"message": "ok"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			err := service.DecryptRequest(t.Context(), data, &test.dst)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, test.dst)
			assert.Empty(t, data.PrivateKeyPEM)
			assert.Empty(t, data.PrivateKeyPKCS8)

			secretManager.AssertExpectations(t)
			cryptoProvider.AssertExpectations(t)
		})
	}
}
