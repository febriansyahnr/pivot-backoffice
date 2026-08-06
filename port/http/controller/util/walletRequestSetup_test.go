package httpControllerUtil

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/stretchr/testify/assert"
)

func TestPrepareInternalWalletRequest(t *testing.T) {
	merchantId := uuid.NewString()
	walletConfig := config.WalletBackendConfig{
		InternalPrefixUrl: "/internal/v1/",
		Host:              "http://example.com",
	}

	tests := []struct {
		name          string
		request       *http.Request
		config        *config.WalletBackendConfig
		expectedError error
		expectedURL   string
	}{
		{
			name: "Valid request with params",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/api/v1/wallet/customers?param=params", nil)
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantId,
				})
				return req.WithContext(ctx)
			}(),
			expectedError: nil,
			expectedURL:   "http://example.com/internal/v1/customers?param=params",
		},
		{
			name: "Valid request",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/api/v1/wallet/customers", nil)
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantId,
				})
				return req.WithContext(ctx)
			}(),
			expectedError: nil,
			expectedURL:   "http://example.com/internal/v1/customers",
		},
		{
			name: "Unauthorized user",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "http://example.com/api/v1/wallet", nil)
				return req
			}(),
			expectedError: pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound),
			expectedURL:   "",
		},
		{
			name: "Error parsing URL",
			request: func() *http.Request {
				req, _ := http.NewRequest("GET", "https://example.com/api/v1/wallet/customers?params=qwerty", nil)
				ctx := context.WithValue(context.Background(), constant.CtxUserInfoKey, &userModel.UserTokenClaims{
					MerchantId: merchantId,
				})
				req = req.WithContext(ctx)
				return req
			}(),
			config: &config.WalletBackendConfig{
				Host: "httpsxL://localhost:3000",
			},
			expectedError: pkgErr.New(response.HttpErrInternal, errors.New("error parse url")),
			expectedURL:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logger.NewSlogger(logger.Config{})

			walletBEConfig := walletConfig
			if tt.config != nil {
				walletBEConfig = *tt.config
			}
			s := &InternalWalletRequestSetup{
				secret: &config.Secret{
					WalletBackendSecret: config.WalletBackendSecret{
						InternalServiceKey: "secret-key",
					},
				},
				config: &config.Config{
					WalletBackendConfig: walletBEConfig,
				},
				logger: logger,
			}

			err := s.PrepareInternalWalletRequest(tt.request)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, tt.request.URL.String())
				assert.Equal(t, "secret-key", tt.request.Header.Get("X-Internal-Service-Key"))
				assert.Equal(t, merchantId, tt.request.Header.Get(constant.HeaderXMerchantId))
			}
		})
	}
}
