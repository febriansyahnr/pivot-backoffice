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

func TestGetSnapPrivateKey(t *testing.T) {
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
			desc:       "error parse pkcs8 private key",
			wantErr:    true,
			merchantID: uuid.NewString(),
			mockSetup: func() {
				invalidPrivate := []byte("-----BEGIN PUBLIC KEY-----\nMIIBCgKCAQEAuw43hAPdhnLpIZ0EsK8ZxCEKg5GhQ8nCOKlNppTsFdzyIgjUni+s\nY6GEdm1eClnIVOomXLvPicllVLJDtno8ezem2faLp/1SO0TItzcytGSt6ucqVia8\ni0Pdc5B4aFF7Fp3ZFLq+fHBA/y6VLQYv5ZrWyb1OrWrxDkibDIOSFJ4zhYiRg7pe\nR7yq23kRRSEEDyWoJ2+S0DwfzEjdeg2AwF+9i0SbN16Z6tXzlLPDDm2kyTRkrW16\n9aV1Cfk7WVP/SwF5AZOMz2sJSN1zft011EPch/n/1UodFVbGSDBTMGuJyaVUx1wl\nbXNmrTxy6xnm9fFuN0y43Impl4Opzd+8pwIDAQAB\n-----END PUBLIC KEY-----\n")
				encMock.On("Decrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&invalidPrivate, nil).Once()
			},
		},
		{
			desc:       "success get snap private key",
			wantErr:    false,
			merchantID: uuid.NewString(),
			mockSetup: func() {
				validPrivate := []byte("-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDI/t4aI3sTFTTH\nLxpdd1HFqKoMUyeHFa9Oiq//OGlMODsR5JHsF+IuRfar/U8HXe6jg0j5sLp/5D8a\ni5MiqvyXJxd1j8EuKKs/qeQrtS0usqqJYJDMze3YX1eIHrD5LO2CBiJhPp6gIhUO\nSfhImxJEMaMDiounZmoK9P66JrEKjedRbuVgDn4PtxSLr3VzYYRZPa65D5HzjdIm\nOuotubMN3mbWRWhicT2oM29fTe3cJlzRAcaOmicPdRtNTf0/Wk+jCDeiJxoLuKyZ\n/IUSVuzIl4gnMGKM/KheSPOlTRwVm+eHWjr2+xB4KZf00FNoJTu+TXUEe0pdYVfw\nioeavUHdAgMBAAECggEBAJNI8EgHJ/Db4Uj0Y0WKYgmNhs5xQM3kPgo35rAHDmIj\n8mUyMRvohH2UFyYBASBM3MpFMfyGXKPLBdLV5IPK+D1rD+294bmJY7PLMsA0i19k\n3UK92F27qUac1u+QTe7J1WEqTZck4+hEEVnfKmlJ+SCvntzBcYTBr4NH9EFEiQdJ\nl1fMfMxS6HgpIcY2wxnypqjpeC/LdG4543iGpP8zbnunxW43yFgkNC8AJ/KuY9eL\ndX/bES25F0mmhQRruzpFczre+4S2tiVhMq73VuyXJiYIvaCxA1JHQjQBCjmkJ7r7\nlNKZYlf7Qfcg5FZ9njGdFWpDeM40vyiGYZbKVV6X4aECgYEA9n84ldyPD4U8Kmnz\n56pJ5lK+2ztC90pXCNS+vpCqg3VdLf7eE1nUAp2f6JBA/qm3kMIfju49bpTFM6ZD\nKj4jI8WI7JdBy1gfNls9HF5nVtoHzTHG9E3wT0YAOPHmyCfohpfFIFaWX63VwZ15\n/jKP9OqW7aeMLoOWv9Mg+WQbVrMCgYEA0L6TKuttEZmyLDFhb+4T52Yl8ofNLbxe\n6BhqDy44xwjBsi2a0BDc9l1/kv2ACsHEfMNFO/jW1IZ4C+2S3WDSoH+ff7uFVxS0\n1wGrNNM/ch/zX3oLzX+ZevAYSys95IshMQMfoePp533ZYk504nRsxM4JVGSUoSHK\nMvmnlRMJzS8CgYEAiJOE7sP+IENaSsXZ9opL1+oRBbeYKxxtjN8TsNLHJ39n2YxV\nz7L93VUovNrwqCmxI+vrQG6QayzS9wMwQ7+aCL/yVeSY9+ojoSJ8gbNs3pp/qBnk\neoiUldfbV7HwhQZXt/tvpbNULj9LKLPwW//382PnrFYhPcR7Sl3Y71WgMDECgYBv\nDzXVe/RHjPJSuOMSXiSQ1LQT2VS8pKAJ9BNZiEoE+w+y8LiRQqeNHCmn1t+s2XLk\nvi+zvKzv3as5DWk6By2I3t3JY8eJkSa1zdl8/XegDIe7oH9vEhhiZCNIuvTvB2bd\nYMAPrebglwB1YTCm2zKTcttb3zeEkym0/Ua/9aUdWQKBgQC43+cW1iWcw3U8LhO7\n3Wkys/d07xkZEoEGJRFymqhyUNIBRY3fe4ukQOp3Qq0bff0Oj5YBDD6n+TES9nSh\nwrKPt/kduqwqr9Ob4SwaUwX18rQlQwRxo2O3EQLcIIMiMVSuDKdc68AT8uc1Dbqn\ngmyBqtD/AVn+rKit0f7HDuOrcw==\n-----END PRIVATE KEY-----\n")
				encMock.On("Decrypt", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(&validPrivate, nil)
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			tc.mockSetup()
			svc := New(merchantRepo, loggerMock, nil, nil, nil, encMock, WithAccountService(accountSvc))
			_, err := svc.GetSnapPrivateKey(context.Background(), tc.merchantID)

			if tc.wantErr {
				assert.Error(t, err)
			} else if err != nil {
				assert.NoError(t, err)
			}
		})
	}
}
