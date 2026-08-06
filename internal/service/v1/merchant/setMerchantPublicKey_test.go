package merchant

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockEncrypt "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/encryption"
	mockMerchant "github.com/paper-indonesia/pivot-backoffice/mocks/repository"
	mocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	mockLogger "github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetMerchantPublicKey(t *testing.T) {
	merchantRepo := mockMerchant.NewIMerchantRepository(t)
	loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	encMock := mockEncrypt.NewICrypto(t)
	accountSvc := mocks.NewIAccountService(t)

	validMerchantId := uuid.NewString()

	testCases := []struct {
		desc            string
		mockSetup       func()
		wantErr         bool
		merchantID      string
		encryptedPubKey string
	}{
		{
			desc:       "error when get merchant auths by id",
			wantErr:    true,
			merchantID: validMerchantId,
			mockSetup: func() {
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(nil, errors.New("error")).Once()
			},
		},
		{
			desc:       "error parse uuid",
			wantErr:    true,
			merchantID: "merchant-id",
			mockSetup: func() {
				merchantRepo.On("GetMerchantSnapPKCS8KeyByMerchantID", mock.Anything, mock.Anything).Return(&merchant.MerchantAuth{
					UUID:       uuid.NewString(),
					MerchantID: uuid.NewString(),
				}, nil)
			},
		},
		{
			desc:       "error when decrypt key",
			wantErr:    true,
			merchantID: uuid.NewString(),
			mockSetup: func() {
				encMock.On("SecretKeyFromUUID", mock.Anything).Return("secret key")
				encMock.On("Decrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil, errors.New("error")).Once()
			},
		},
		{
			desc:       "error invalid publicKey",
			wantErr:    true,
			merchantID: uuid.NewString(),
			mockSetup: func() {
				invalidPublicKey := []byte("invalid public key")
				encMock.On("Decrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&invalidPublicKey, nil).Once()
			},
		},
		{
			desc:       "error when update merchant auth",
			wantErr:    true,
			merchantID: uuid.NewString(),
			mockSetup: func() {
				validPublic := []byte("-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEApdgUB+pG0ywUJ4phFWcm\nhJdTMbbY7j5ZJCY1jhhm+79M2qCEL+Qn0ju3xy6eZjJTbEtIosEJy9xeoQOjEKBq\nBIZSOFeU2Mk06JvxyXCJo2VJWp8F+Rpt8Y7aCuyinyVuvO8vEdONOXr2l0m4Snwt\nLXuWU6/u3WjgaQqleLbc8UBsE0IqqrLPJeWlNUq5wi7/oD1mSdTGMGBIkeje4+N0\nKvWDoGl086r9BleWKwsbkPu8zSKJZ/4c7F3OB/SXWn8TvBv3VvgEMDD6ikhGZ8oa\ndfxpF3shp4E4bge9jsFJLBbQn8MQTK3EN6knvTc5HYDfmEQlWRqx2BgBUw8GEWUa\nhQIDAQAB\n-----END PUBLIC KEY-----\n")
				encMock.On("Decrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&validPublic, nil)
				merchantRepo.On("UpdateMerchantAuth", mock.Anything, mock.Anything).Return(errors.New("error")).Once()
			},
		},
		{
			desc: "success set merchant public key",
			mockSetup: func() {
				merchantRepo.On("UpdateMerchantAuth", mock.Anything, mock.Anything).Return(nil)
			},
			merchantID:      uuid.NewString(),
			encryptedPubKey: "something data encrypted",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.mockSetup()
			svc := New(merchantRepo, loggerMock, nil, nil, nil, encMock, WithAccountService(accountSvc))
			err := svc.SetMerchantPublicKey(context.Background(), tc.merchantID, tc.encryptedPubKey)

			if tc.wantErr {
				assert.Error(t, err)
			} else if err != nil {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUtilEncryptingKey(t *testing.T) {
	encMock := mockEncrypt.NewICrypto(t)
	loggerMock, _ := mockLogger.NewZapLogger(mockLogger.Config{})
	accountSvc := mocks.NewIAccountService(t)

	testCases := []struct {
		desc      string
		input     func() (key string, data string)
		mockSetup func()
		wantErr   bool
	}{
		{
			desc: "error parse id and length is not 32",
			input: func() (key string, data string) {
				return "test key", "test data"
			},
			mockSetup: func() {
			},
			wantErr: true,
		},
		{
			desc:    "error encrypt and key is not uuid and length is 32",
			wantErr: true,
			input: func() (key string, data string) {
				return "111111111122222222223333333333*&", "test data"
			},
			mockSetup: func() {
				encMock.On("Encrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", errors.New("error")).Once()
			},
		},
		{
			desc: "success encrypt and key length is 32",
			input: func() (key string, data string) {
				return "111111111122222222223333333333*&", "test data"
			},
			mockSetup: func() {
				encMock.On("Encrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("something encryption", nil).Once()
			},
		},
		{
			desc: "error encrypt and key is uuid",
			input: func() (key string, data string) {
				return uuid.NewString(), "test data"
			},
			mockSetup: func() {
				encMock.On("SecretKeyFromUUID", mock.Anything).Return("111111111122222222223333333333*&")
				encMock.On("Encrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("", errors.New("error")).Once()
			},
			wantErr: true,
		},
		{
			desc: "success encrypt and key is uuid",
			input: func() (key string, data string) {
				return uuid.NewString(), "test data"
			},
			mockSetup: func() {
				encMock.On("Encrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return("something encryption", nil)
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			tC.mockSetup()

			svc := New(nil, loggerMock, nil, nil, nil, encMock, WithAccountService(accountSvc))

			key, data := tC.input()

			res, err := svc.UtilEncryptingKey(context.Background(), key, data)

			if tC.wantErr {
				assert.Error(t, err)
			} else if err != nil {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}
