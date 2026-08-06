package merchant

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/google/uuid"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMerchantServiceGetPKCS8SecretKey(t *testing.T) {

	vaultTransit := vaultMock.NewIVaultTransit(t)

	merchantAuth := merchant.MerchantAuth{
		UUID:       uuid.NewString(),
		MerchantID: uuid.NewString(),
		Secret:     "secret",
		MerchantPublicKey: sql.NullString{
			String: "public",
			Valid:  true,
		},
		SnapPrivateKey: sql.NullString{
			String: "private",
			Valid:  true,
		},
		CreatedAt: time.Now(),
	}

	testCases := []struct {
		desc      string
		input     string
		mockSetup func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto)
		wantErr   bool
	}{
		{
			desc:  "SUCCESS: Get PKCS8 Secret Key",
			input: merchantAuth.UUID,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(&merchantAuth, nil)
			},
		},
		{
			desc:  "SUCCESS: Get PKCS8 Secret Key With Encrypted Secret",
			input: merchantAuth.UUID,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, _ *mockEncrypt.ICrypto) {
				result := merchantAuth
				result.Secret = "vault:v1:..."
				result.SecretVersion = 1
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(&result, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(&vault.DecryptResponse{Plaintext: []byte(`secret`)}, nil)
			},
		},
		{
			desc:  "ERROR: Get PKCS8 Secret Key",
			input: merchantAuth.UUID,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(nil, errors.New("error get merchant auth"))
			},
			wantErr: true,
		},
		{
			desc:  "ERROR: Get PKCS8 Secret Key With Encrypted Secret",
			input: merchantAuth.UUID,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, _ *mockEncrypt.ICrypto) {
				result := merchantAuth
				result.Secret = "vault:v1:..."
				result.SecretVersion = 1
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(&result, nil)
				vaultTransit.On("Decrypt", mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantErr: true,
		},
		{
			desc:  "SUCCESS: Get PKCS8 Secret Key with empty SnapPrivateKey",
			input: merchantAuth.UUID,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				result := merchantAuth
				result.SnapPrivateKey = sql.NullString{}
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(&result, nil)
			},
			wantErr: false,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trxRepoMock := mockMerchant.NewIMerchantRepository(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			encMock := mockEncrypt.NewICrypto(t)
			accountSvc := mocks.NewIAccountService(t)

			tc.mockSetup(trxRepoMock, encMock)

			svc := New(trxRepoMock, loggerMock, nil, nil, nil, encMock, WithAccountService(accountSvc), WithVaultTransit(vaultTransit))

			result, err := svc.GetPKCS8SecretKey(context.Background(), tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("GetPKCS8SecretKey() error = %v, wantErr %v", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Errorf("GetPKCS8SecretKey() error = %v, wantErr %v", err, tc.wantErr)
				}
				if result == nil {
					t.Errorf("GetPKCS8SecretKey() result = %v, wantErr %v", result, tc.wantErr)
				}
			}
		})
	}
}
