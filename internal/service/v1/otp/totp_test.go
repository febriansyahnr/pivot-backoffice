package otp_test

import (
	"crypto/aes"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	. "github.com/paper-indonesia/pivot-backoffice/internal/service/v1/otp"
	vaultMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/vault"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pivot-backoffice/pkg/vault"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestValidateTOTPCode(t *testing.T) {
	encryptionKey := vaultMock.NewIVaultKeyValue(t)

	config := &config.Config{
		MultiFactorAuth: config.MultiFactorAuthConfig{
			TimeBasedOTP: config.TimeBasedOTPConfig{
				TOTPIssuer: "PivotPayment", // NOSONAR
			},
		},
	}
	service := New(config, nil, nil, nil, nil, nil, nil)
	service.WithUserEncryptionKey(encryptionKey)

	totpSecretKey := "QM3UAJ6KP6NRYJTJR43MIV3PXJLC6ZPU7YCJHHSEPV7QL2PMB6DA" // NOSONAR
	request := &otp.VerifyTOTPRequest{
		EncryptVersion: util.ValueToPtr(1),                                                                                                              // NOSONAR
		WrappedSecret:  util.ValueToPtr("omlghZWihCisq4LjzdiIDpipSwj1V4EuH8UPRVI9jwZF+mIev1Eih3PVGySKsipnLCJ7Yj/MdJlLEZe6n/LitBfOBJ2ybFYTQwTy9kapivs="), // NOSONAR
	}

	tests := []struct {
		name         string
		request      *otp.VerifyTOTPRequest
		generateCode func(t *testing.T) string
		setupMock    func()
		wantError    error
		wantResult   bool
	}{
		{
			name:       "ERROR:Encrypt version is nil",
			request:    &otp.VerifyTOTPRequest{},
			wantError:  pkgErrs.New(response.HttpErrUnprocessableContent, errors.New("user has not registered for TOTP")),
			wantResult: false,
		},
		{
			name:    "ERROR:Get secret version",
			request: request,
			setupMock: func() {
				encryptionKey.On("GetSecretKeyVersionString", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, assert.AnError)
			},
			wantError:  pkgErrs.New(response.HttpErrInternal, assert.AnError),
			wantResult: false,
		},
		{
			name:    "ERROR:Invalid secret value format",
			request: request,
			setupMock: func() {
				encryptionKey.On("GetSecretKeyVersionString", mock.Anything, mock.Anything, mock.Anything).Once().Return(&vault.SecretKey[string]{Value: "123456"}, nil)
			},
			wantError:  pkgErrs.New(response.HttpErrInternal, base64.CorruptInputError(4)),
			wantResult: false,
		},
		{
			name:    "ERROR:Invalid encryption key",
			request: request,
			setupMock: func() {
				encryptionKey.On("GetSecretKeyVersionString", mock.Anything, mock.Anything, mock.Anything).Once().Return(&vault.SecretKey[string]{Value: ""}, nil)
			},
			wantError:  pkgErrs.New(response.HttpErrInternal, aes.KeySizeError(0)),
			wantResult: false,
		},
		{
			name:    "SUCCESS:Invalid TOTP code (random code)",
			request: request,
			setupMock: func() {
				encryptionKey.On("GetSecretKeyVersionString", mock.Anything, mock.Anything, mock.Anything).Return(&vault.SecretKey[string]{Value: "U9JqU15dTbHFDhnvTYCYmvcbrFcm1oHvhq/WXW6cPcI="}, nil) // NOSONAR
			},
			wantError: nil, wantResult: false,
		},
		{
			name:    "SUCCESS:Invalid TOTP code (backtime code)",
			request: request,
			generateCode: func(t *testing.T) string {
				code, err := totp.GenerateCode(totpSecretKey, time.Now().Add(-60*time.Second))
				require.NoError(t, err)

				return code
			},
			wantError: nil, wantResult: false,
		},
		{
			name:    "SUCCESS:Valid TOTP code",
			request: request,
			generateCode: func(t *testing.T) string {
				code, err := totp.GenerateCode(totpSecretKey, time.Now().Add(-5*time.Second))
				require.NoError(t, err)

				return code
			},
			wantError: nil, wantResult: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}
			if test.generateCode == nil {
				test.generateCode = func(_ *testing.T) string { return "123456" }
			}
			test.request.Code = test.generateCode(t)

			result, err := service.ValidateTOTPCode(t.Context(), test.request)
			assert.Equal(t, test.wantError, err)
			assert.Equal(t, test.wantResult, result)
		})
	}
}
