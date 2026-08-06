package merchant

import (
	"context"
	"database/sql"
	"testing"
	"time"

	errors "errors"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	responseHttp "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/mock"
)

func TestMerchantService_CreatePKCS8SecretKey(t *testing.T) {
	merchantAuth := &merchant.MerchantAuth{
		UUID:       uuid.NewString(),
		Secret:     "secret",
		MerchantID: uuid.NewString(),
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
			desc:    "ERROR: merchant auth not found",
			input:   uuid.NewString(),
			wantErr: true,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(nil, pkgErrors.New(responseHttp.HttpErrNotFound, errors.New("merchant auth not found")))
			},
		},
		{
			desc:    "ERROR: failed to generate pkcs8 secret key",
			input:   uuid.NewString(),
			wantErr: true,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(merchantAuth, nil)
				enc.On("GenerateRandomPKCS8Key").Return(nil, pkgErrors.New(responseHttp.HttpErrInternal, errors.New("failed to generate pkcs8 secret key")))
			},
		},
		{
			desc:    "ERROR: failed to encrypt secret key",
			input:   uuid.NewString(),
			wantErr: true,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(merchantAuth, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("secret key from merchant")
				enc.On("GenerateRandomPKCS8Key").Return([]byte(merchantAuth.SnapPrivateKey.String), nil)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("", errors.New("failed to encrypt key"))
			},
		},
		{
			desc:    "ERROR: failed to update secret key",
			input:   uuid.NewString(),
			wantErr: true,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(merchantAuth, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("secret key from merchant")
				enc.On("GenerateRandomPKCS8Key").Return([]byte(merchantAuth.SnapPrivateKey.String), nil)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("some secret", nil)
				trxRepo.On("UpdateMerchantAuth", mock.Anything, mock.Anything).Return(errors.New("failed to update secret key"))
			},
		},
		{
			desc:    "ERROR: failed to parse merchantID",
			input:   "invalid-uuid",
			wantErr: true,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(merchantAuth, nil)
				enc.On("GenerateRandomPKCS8Key").Return([]byte(merchantAuth.SnapPrivateKey.String), nil)
			},
		},
		{
			desc:    "SUCCESS: create pkcs8 secret key",
			input:   uuid.NewString(),
			wantErr: false,
			mockSetup: func(trxRepo *mockMerchant.IMerchantRepository, enc *mockEncrypt.ICrypto) {
				trxRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(merchantAuth, nil)
				enc.On("SecretKeyFromUUID", mock.Anything).Return("secret key from merchant")
				enc.On("GenerateRandomPKCS8Key").Return([]byte(merchantAuth.SnapPrivateKey.String), nil)
				enc.On("Encrypt", mock.Anything, mock.Anything).Return("some secret", nil)
				trxRepo.On("UpdateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			trxRepo := mockMerchant.NewIMerchantRepository(t)
			encMock := mockEncrypt.NewICrypto(t)
			loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
			accountSvc := mocks.NewIAccountService(t)
			tc.mockSetup(trxRepo, encMock)

			svc := New(trxRepo, loggerMock, nil, nil, nil, encMock, WithAccountService(accountSvc))
			result, err := svc.CreatePKCS8SecretKey(context.Background(), tc.input)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if result != nil {
					t.Errorf("result expected nil, got %v", result)
				}
			} else {
				t.Logf("result %v", result)

				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
				if result == nil {
					t.Errorf("expected result, got nil")
				}
			}

			trxRepo.AssertExpectations(t)
			encMock.AssertExpectations(t)

		})
	}
}
