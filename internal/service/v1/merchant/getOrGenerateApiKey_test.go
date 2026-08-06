package merchant

import (
	"context"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetOrGenerateCallbackApiKey(t *testing.T) {
	merchantID := uuid.NewString()

	vaultTransit := vaultMock.NewIVaultTransit(t)

	testCases := []struct {
		Name      string
		IsSuccess bool
		MockSetup func(mockRepo *mockMerchant.IMerchantRepository)
	}{
		{
			Name:      "SUCCESS: Plaintext callback api key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:           merchantID,
					CallbackApiKey: util.ValueToPtr("api-key"),
				}, nil)
			},
		},
		{
			Name:      "SUCCESS: Unwrapped callback api key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:                  merchantID,
					CallbackApiKey:        util.ValueToPtr("vault:v1:..."),
					CallbackApiKeyVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{Plaintext: []byte(`api-key`)}, nil)
			},
		},
		{
			Name:      "SUCCESS: Regenerate callback api key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(&vault.EncryptResponse{}, nil)
				mockRepo.On(
					"UpdateCallbackApiKey", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything,
				).Return(nil)
			},
		},
		{
			Name:      "ERROR: Find merchant",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Merchant data not found",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On("FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType()).Return(nil, nil)
			},
		},
		{
			Name:      "ERROR: Decrypt callback api key",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:                  merchantID,
					CallbackApiKey:        util.ValueToPtr("vault:v1:..."),
					CallbackApiKeyVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Encrypt new callback api key",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Update new callback api key",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(&vault.EncryptResponse{}, nil)
				mockRepo.On(
					"UpdateCallbackApiKey", constant.ValueCtxMockType(), constant.StringMockType(), constant.StringMockType(), mock.Anything,
				).Return(assert.AnError)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := mockMerchant.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			accountSvc := mocks.NewIAccountService(t)

			tc.MockSetup(merchantRepo)
			svc := New(merchantRepo, loggerMock, nil, nil, nil, nil, WithAccountService(accountSvc), WithVaultTransit(vaultTransit))

			response, err := svc.GetOrGenerateCallbackApiKey(t.Context(), merchantID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}

func TestGetOrGenerateJITApiKey(t *testing.T) {
	merchantID := uuid.NewString()
	vaultTransit := vaultMock.NewIVaultTransit(t)

	testCases := []struct {
		Name      string
		IsSuccess bool
		MockSetup func(
			mockRepo *mockMerchant.IMerchantRepository,

		)
	}{
		{
			Name:      "SUCCESS: Plaintext JIT API key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:      merchantID,
					JITApiKey: "api-key",
				}, nil)
			},
		},
		{
			Name:      "SUCCESS: Unwrapped JIT API key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID:             merchantID,
					JITApiKey:        "vault:v1:...",
					JITApiKeyVersion: 1,
				}, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{Plaintext: []byte(`api-key`)}, nil)
			},
		},
		{
			Name:      "SUCCESS: Regenerate JIT API key",
			IsSuccess: true,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(&vault.EncryptResponse{}, nil)
				mockRepo.On("Update", constant.ValueCtxMockType(), mock.Anything).Return(nil)
			},
		},
		{
			Name:      "ERROR: Find merchant",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Merchant not found",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(nil, nil)
			},
		},
		{
			Name:      "ERROR: Encrypt new JIT API key",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
		},
		{
			Name:      "ERROR: Update merchant data",
			IsSuccess: false,
			MockSetup: func(mockRepo *mockMerchant.IMerchantRepository) {
				mockRepo.On(
					"FindMerchantByID", constant.ValueCtxMockType(), constant.StringMockType(),
				).Return(&merchant.Merchant{
					UUID: merchantID,
				}, nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(&vault.EncryptResponse{}, nil)

				mockRepo.On(
					"Update", constant.ValueCtxMockType(), mock.Anything,
				).Return(assert.AnError)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			merchantRepo := mockMerchant.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			accountSvc := mocks.NewIAccountService(t)

			tc.MockSetup(merchantRepo)
			svc := New(merchantRepo, loggerMock, nil, nil, nil, nil, WithAccountService(accountSvc), WithVaultTransit(vaultTransit))

			response, err := svc.GetOrGenerateJITApiKey(context.Background(), merchantID)
			if tc.IsSuccess {
				assert.NoError(t, err)
				require.NotEmpty(t, response)
			} else {
				require.Error(t, err)
				require.Empty(t, response)
			}

			merchantRepo.AssertExpectations(t)
		})
	}
}
