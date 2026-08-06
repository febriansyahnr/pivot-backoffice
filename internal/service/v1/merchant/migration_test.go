package merchant_test

import (
	"errors"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/merchant"
	loggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repoMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMigrateMerchantSecretsToEncryption(t *testing.T) {
	logger := loggerMock.NewILogger(t)
	vaultTransit := vaultMock.NewIVaultTransit(t)
	merchantRepo := repoMocks.NewIMerchantRepository(t)

	service := New(merchantRepo, logger, nil, nil, nil, nil, WithVaultTransit(vaultTransit))

	merchants := []merchant.UnencryptedMerchantSecretsForMigration{
		{
			MerchantId:     "f99fb79b-ca61-438d-a363-27e358e4fa0e", // NOSONAR
			CallbackApiKey: "abc",                                  // NOSONAR
			JITApiKey:      "def",                                  // NOSONAR
			Secret:         "ghijkl",                               // NOSONAR
		},
	}
	batchEncryptInput := vault.BatchEncryptRequest{
		BatchInput: []vault.BatchEncryptInput{
			{Plaintext: []byte(`abc`)}, {Plaintext: []byte(`def`)}, {Plaintext: []byte(`ghijkl`)}, // NOSONAR
		},
	}
	encryptionResults := []vault.EncryptResponse{
		{Ciphertext: "vault:v1:abc...", KeyVersion: 1}, // NOSONAR
		{Ciphertext: "vault:v1:def...", KeyVersion: 1}, // NOSONAR
		{Ciphertext: "vault:v1:ghi...", KeyVersion: 1}, // NOSONAR
	}
	migrateMerchantSecretsRequest := merchant.MigrateMerchantSecretsToEncryption{
		MerchantId:            merchants[0].MerchantId, // NOSONAR
		CallbackApiKey:        "vault:v1:abc...",       // NOSONAR
		CallbackApiKeyVersion: 1,                       // NOSONAR
		JITApiKey:             "vault:v1:def...",       // NOSONAR
		JITApiKeyVersion:      1,                       // NOSONAR
		Secret:                "vault:v1:ghi...",       // NOSONAR
		SecretVersion:         1,                       // NOSONAR
	}
	tests := []struct {
		name      string
		setupMock func()
		wantError error
	}{
		{
			name: "ERROR:List unencrypted merchant secrets",
			setupMock: func() {
				merchantRepo.On("ListUnencryptedMerchantSecretsForMigration", mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError: assert.AnError,
		},
		{
			name: "SUCCESS:No data will be migrated",
			setupMock: func() {
				merchantRepo.On("ListUnencryptedMerchantSecretsForMigration", mock.Anything).Once().Return(nil, nil)
				logger.On("Info", mock.Anything, "No data will be migrated, the process is stopped").Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Batch encrypt merchant secrets",
			setupMock: func() {
				merchantRepo.On("ListUnencryptedMerchantSecretsForMigration", mock.Anything).Return(merchants, nil)
				vaultTransit.On("BatchEncrypt", mock.Anything, batchEncryptInput).Once().Return(nil, assert.AnError)
				logger.On(
					"Info", mock.Anything, "Starting the merchant secret encryption process", pdkLog.Int("total", 1),
				).Once().Return()
				logger.On(
					"Info", mock.Anything, "Secret encryption process for merchant ID: f99fb79b-ca61-438d-a363-27e358e4fa0e", pdkLog.Bool("completed", false), pdkLog.Error(assert.AnError),
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Encryption results do not match",
			setupMock: func() {
				vaultTransit.On("BatchEncrypt", mock.Anything, batchEncryptInput).Once().Return(nil, nil)
				logger.On(
					"Info", mock.Anything, "Starting the merchant secret encryption process", pdkLog.Int("total", 1),
				).Once().Return()
				logger.On(
					"Info", mock.Anything, "Secret encryption process for merchant ID: f99fb79b-ca61-438d-a363-27e358e4fa0e", pdkLog.Bool("completed", false), pdkLog.Error(errors.New("encryption results do not match, result 0 should be 3")),
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "ERROR:Migrate merchant secrets to encryption",
			setupMock: func() {
				vaultTransit.On("BatchEncrypt", mock.Anything, batchEncryptInput).Return(encryptionResults, nil)
				merchantRepo.On("MigrateMerchantSecretsToEncryption", mock.Anything, migrateMerchantSecretsRequest).Once().Return(assert.AnError)
				logger.On(
					"Info", mock.Anything, "Starting the merchant secret encryption process", pdkLog.Int("total", 1),
				).Once().Return()
				logger.On(
					"Info", mock.Anything, "Secret encryption process for merchant ID: f99fb79b-ca61-438d-a363-27e358e4fa0e", pdkLog.Bool("completed", false), pdkLog.Error(assert.AnError),
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS", // NOSONAR
			setupMock: func() {
				merchantRepo.On("MigrateMerchantSecretsToEncryption", mock.Anything, migrateMerchantSecretsRequest).Once().Return(nil)
				logger.On(
					"Info", mock.Anything, "Starting the merchant secret encryption process", pdkLog.Int("total", 1),
				).Once().Return()
				logger.On(
					"Info", mock.Anything, "Secret encryption process for merchant ID: f99fb79b-ca61-438d-a363-27e358e4fa0e", pdkLog.Bool("completed", true), pdkLog.Error(nil),
				).Once().Return()
			},
			wantError: nil,
		},
		{
			name: "SUCCESS:Empty JIT API Key", // NOSONAR
			setupMock: func() {
				merchants[0].JITApiKey = ""
				batchEncryptInput.BatchInput[1].Plaintext = []byte("")
				migrateMerchantSecretsRequest.JITApiKey = ""
				migrateMerchantSecretsRequest.JITApiKeyVersion = 0

				merchantRepo.On("MigrateMerchantSecretsToEncryption", mock.Anything, migrateMerchantSecretsRequest).Once().Return(nil)
				logger.On(
					"Info", mock.Anything, "Starting the merchant secret encryption process", pdkLog.Int("total", 1),
				).Once().Return()
				logger.On(
					"Info", mock.Anything, "Secret encryption process for merchant ID: f99fb79b-ca61-438d-a363-27e358e4fa0e", pdkLog.Bool("completed", true), pdkLog.Error(nil),
				).Once().Return()
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantError, service.MigrateMerchantSecretsToEncryption(t.Context()))

			logger.AssertExpectations(t)
			merchantRepo.AssertExpectations(t)
			vaultTransit.AssertExpectations(t)
		})
	}
}
