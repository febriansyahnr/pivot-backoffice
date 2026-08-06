package merchant

import (
	"context"
	"errors"
	"testing"

	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	repositoryMocks "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateMerchantAuth(t *testing.T) {

	vaultTransit := vaultMock.NewIVaultTransit(t)

	testCases := []struct {
		name       string
		mocksSetup func(enc *mockEncrypt.ICrypto, merchRepo *repositoryMocks.IMerchantRepository)
		wantErr    bool
	}{
		{
			name: "ERROR: generate random pkcs8 key",
			mocksSetup: func(enc *mockEncrypt.ICrypto, _ *repositoryMocks.IMerchantRepository) {
				enc.On("GenerateRandomPKCS8Key").Return(nil, errors.New("test"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Encrypt private & secret key",
			mocksSetup: func(enc *mockEncrypt.ICrypto, _ *repositoryMocks.IMerchantRepository) {
				enc.On("GenerateRandomPKCS8Key").Return(nil, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("nil", errors.New("error"))
			},
			wantErr: true,
		},
		{
			name: "ERROR: Encrypt merchant secret",
			mocksSetup: func(enc *mockEncrypt.ICrypto, _ *repositoryMocks.IMerchantRepository) {
				enc.On("GenerateRandomPKCS8Key").Return(nil, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "ERROR: Create Merchant Auth",
			mocksSetup: func(enc *mockEncrypt.ICrypto, merchRepo *repositoryMocks.IMerchantRepository) {
				enc.On("GenerateRandomPKCS8Key").Return(nil, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)
				vaultTransit.On("Encrypt", mock.Anything, mock.Anything).Return(&vault.EncryptResponse{}, nil)

				merchRepo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			wantErr: false,
		},
		{
			name: "SUCCESS: create merchant auth",
			mocksSetup: func(enc *mockEncrypt.ICrypto, merchRepo *repositoryMocks.IMerchantRepository) {
				enc.On("GenerateRandomPKCS8Key").Return(nil, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("nil")
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("nil", nil)

				merchRepo.On("CreateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			merchantRepo := repositoryMocks.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			encryptMock := mockEncrypt.NewICrypto(t)

			tc.mocksSetup(encryptMock, merchantRepo)

			svc := New(merchantRepo, loggerMock, nil, nil, nil, encryptMock, WithVaultTransit(vaultTransit))

			err := svc.CreateMerchantAuth(context.Background(), "merchantId")
			if tc.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}
