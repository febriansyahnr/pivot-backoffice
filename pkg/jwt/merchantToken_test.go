package jwt

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJwtConfigVerifyMerchantToken(t *testing.T) {
	testCases := []struct {
		name    string
		token   func() string
		wantErr bool
	}{
		{
			name: "Success",
			token: func() string {
				merchantClaims := &merchant.MerchantAuthTokenClaims{
					ClientId:   uuid.NewString(),
					MerchantId: uuid.NewString(),
				}

				merchantClaims.RegisteredClaims = goJwt.RegisteredClaims{
					Subject:   merchantClaims.ClientId,
					Issuer:    "testing",
					ExpiresAt: goJwt.NewNumericDate(util.TimeNow.Add(constant.LOGIN_EXPIRATION_DURATION)),
				}

				newToken := goJwt.NewWithClaims(goJwt.SigningMethodHS256, merchantClaims)
				token, _ := newToken.SignedString([]byte("testing"))
				return token
			},
			wantErr: false,
		},
		{
			name: "Error: expired token",
			token: func() string {
				merchantClaims := &merchant.MerchantAuthTokenClaims{
					ClientId:   uuid.NewString(),
					MerchantId: uuid.NewString(),
				}

				merchantClaims.RegisteredClaims = goJwt.RegisteredClaims{
					Subject:   merchantClaims.ClientId,
					Issuer:    "testing",
					ExpiresAt: goJwt.NewNumericDate(util.TimeNow.Add(-constant.LOGIN_EXPIRATION_DURATION)),
				}

				newToken := goJwt.NewWithClaims(goJwt.SigningMethodHS256, merchantClaims)
				token, _ := newToken.SignedString([]byte("testing"))
				return token
			},
			wantErr: true,
		},
		{
			name:    "Error: token malformed",
			token:   func() string { return "malformed" },
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantKey: "testing",
				},
			}

			redis := mockRedis.NewIRedisExt(t)

			jwtCore, _ := New(cfg, secret, redis)

			ctx := context.Background()
			_, err := jwtCore.VerifyMerchantToken(ctx, tc.token())

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}
			}
		})
	}
}

func TestJWTConfigGenerateMerchantToken(t *testing.T) {

	pwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(pwd, "backend-portal")
	dir := path.Join(projectRoot, "test", "dummies")

	rsaKeyPair := &config.RSAKeyPairFile{
		PrivateKeyFile: dir + "/valid_id_rsa",
		PublicKeyFile:  dir + "/valid_id_rsa.pub",
	}

	testCases := []struct {
		name       string
		rsaKeyPair *config.RSAKeyPairFile
		wantErr    bool
	}{
		{
			name:    "ERROR: Missing merchant RSA key pair",
			wantErr: true,
		},
		{
			name:       "SUCCESS",
			rsaKeyPair: rsaKeyPair,
			wantErr:    false,
		},
	}

	ctx := context.Background()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantKey:    "abc123!@#",
					MerchantRSAKey: tc.rsaKeyPair,
				},
			}

			redis := mockRedis.NewIRedisExt(t)

			jwtCore, _ := New(cfg, secret, redis)

			token, err := jwtCore.GenerateMerchantToken(ctx, "client-id", "merchant-id")

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				if _, errSign := jwtCore.VerifyMerchantToken(ctx, token); errSign != nil {
					t.Errorf("unexpected error: %s", errSign)
				}
			}
		})
	}
}
