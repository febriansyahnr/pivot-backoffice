package jwt

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path"
	"testing"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/otp"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	mockRedis "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/go-redis/redismock/v9"
	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTConfig_GenerateAccessToken(t *testing.T) {
	expectedUser := &user.User{
		UUID:         "uuid-uuid-uuid",
		Name:         "test",
		Email:        "test@gmail.com",
		Password:     "pass",
		MerchantId:   "merchant-id",
		RefreshToken: sql.NullString{String: "refresh-token", Valid: true},
		Role:         sql.NullString{String: "admin", Valid: true},
		CreatedAt:    time.Now(),
	}

	userClaims := &user.UserTokenClaims{
		UUID:         "uuid-uuid-uuid",
		Name:         "test",
		Email:        "test@gmail.com",
		MerchantId:   "merchant-id",
		RefreshToken: "refresh-token",
	}

	userClaims.RegisteredClaims = goJwt.RegisteredClaims{
		Subject:   userClaims.UUID,
		Issuer:    "testing",
		ExpiresAt: goJwt.NewNumericDate(util.TimeNow.Add(constant.LOGIN_EXPIRATION_DURATION)),
	}

	testCases := []struct {
		name           string
		expectedClaims *user.UserTokenClaims
		wantErr        bool
	}{
		{
			name:           "it should return token",
			expectedClaims: userClaims,
			wantErr:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			redis := mockRedis.NewIRedisExt(t)

			jwtCore, _ := New(cfg, secret, redis)

			ctx := context.Background()
			token, err := jwtCore.GenerateAccessToken(ctx, expectedUser)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				_, errSign := jwtCore.Verify(ctx, token)
				if errSign != nil {
					t.Errorf("unexpected error: %s", errSign)
				}
			}
		})
	}
}

func TestJWTConfig_GenerateRefreshToken(t *testing.T) {
	expectedUser := &user.User{
		UUID:         "uuid-uuid-uuid",
		Name:         "test",
		Email:        "test@gmail.com",
		Password:     "pass",
		MerchantId:   "merchant-id",
		RefreshToken: sql.NullString{String: "refresh-token", Valid: true},
		Role:         sql.NullString{String: "admin", Valid: true},
		CreatedAt:    time.Now(),
	}

	userClaims := &user.UserTokenClaims{
		UUID:         "uuid-uuid-uuid",
		Name:         "test",
		Email:        "test@gmail.com",
		MerchantId:   "merchant-id",
		RefreshToken: "refresh-token",
	}

	userClaims.RegisteredClaims = goJwt.RegisteredClaims{
		Subject:   userClaims.UUID,
		Issuer:    "testing",
		ExpiresAt: goJwt.NewNumericDate(util.TimeNow.Add(constant.LOGIN_EXPIRATION_DURATION)),
	}

	testCases := []struct {
		name           string
		expectedClaims *user.UserTokenClaims
		wantErr        bool
	}{
		{
			name:           "it should return token",
			expectedClaims: userClaims,
			wantErr:        false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := &config.Config{
				ServiceName: "testing",
			}

			secret := &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					UserKey: "testing",
				},
			}

			redis := mockRedis.NewIRedisExt(t)

			jwtCore, _ := New(cfg, secret, redis)

			ctx := context.Background()
			token, err := jwtCore.GenerateRefreshToken(ctx, expectedUser, time.Now().Add(constant.LOGIN_EXPIRATION_DURATION))

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %s", err)
				}

				_, errSign := jwtCore.Verify(ctx, token)
				if errSign != nil {
					t.Errorf("unexpected error: %s", errSign)
				}
			}
		})
	}
}

func TestGenerateTokenForOTP(t *testing.T) {

	idUser := "uuid-user-1"
	newJwt, _ := New(
		&config.Config{}, &config.Secret{JWTSignatureKey: config.JWTSignatureKey{TokenOTPKey: "test-jwt-key-001"}}, nil,
	)
	token, err := newJwt.GenerateTokenForOTP(context.Background(), idUser, constant.OTPIdentifierForgotPassword)
	require.Nil(t, err)

	claims, err := newJwt.ValidateTokenFromOTP(context.Background(), token)
	require.Nil(t, err)
	assert.Equal(t, idUser, claims.UUID)
	assert.Equal(t, constant.OTPIdentifierForgotPassword, claims.Identifier)
}

func TestGenerateTokenForFeature2FA(t *testing.T) {

	idUser := "uuid-user-2"
	newJwt, _ := New(
		&config.Config{}, &config.Secret{JWTSignatureKey: config.JWTSignatureKey{TokenOTPFeatureKey: "test-jwt-key-002"}}, nil,
	)
	token, err := newJwt.GenerateTokenForFeature2FA(context.Background(), idUser, constant.OTPIdentifierForgotPassword)
	require.Nil(t, err)

	claims, err := newJwt.ValidateTokenFromFeature2FA(context.Background(), token)
	require.Nil(t, err)
	assert.Equal(t, idUser, claims.UUID)
	assert.Equal(t, constant.OTPIdentifierForgotPassword, claims.Identifier)
}

func TestJwtConfig_Verify(t *testing.T) {
	testCases := []struct {
		name    string
		token   func() string
		wantErr bool
	}{
		{
			name: "Success",
			token: func() string {
				userClaims := &user.UserTokenClaims{
					UUID:         "uuid-uuid-uuid",
					Name:         "test",
					Email:        "test@gmail.com",
					MerchantId:   "merchant-id",
					RefreshToken: "refresh-token",
				}

				userClaims.RegisteredClaims = goJwt.RegisteredClaims{
					Subject:   userClaims.UUID,
					Issuer:    "testing",
					ExpiresAt: goJwt.NewNumericDate(util.TimeNow.Add(constant.LOGIN_EXPIRATION_DURATION)),
				}

				newToken := goJwt.NewWithClaims(goJwt.SigningMethodHS256, userClaims)
				token, _ := newToken.SignedString([]byte("testing"))
				return token
			},
			wantErr: false,
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
					UserKey: "testing",
				},
			}

			redis := mockRedis.NewIRedisExt(t)

			jwtCore, _ := New(cfg, secret, redis)

			ctx := context.Background()
			_, err := jwtCore.Verify(ctx, tc.token())

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

func TestValidateTokenFromOTP(pt *testing.T) {
	newJwt, _ := New(
		&config.Config{}, &config.Secret{JWTSignatureKey: config.JWTSignatureKey{TokenOTPKey: "test-jwt-key-001"}}, nil,
	)

	tests := []struct {
		name       string
		token      string
		wantErr    string
		wantResult *otpModel.TokenOTPClaims
	}{
		{
			name:    "Invalid token key",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.DMp1LCRvKrG76ZKY99O7N0rrwaEsfbF-hm-jzKNVisY",
			wantErr: "token signature is invalid",
		},
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.awoAo8i2uIvROrmxDqS1aa45N4OOPJU7KHuiTwFaBls",
			wantResult: &otpModel.TokenOTPClaims{
				UUID:       "ce93758e-b19b-43c2-972a-f64be16f08cf",
				Identifier: "31b3d099-8508-4e93-a80c-0471ef237a4c",
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			claims, err := newJwt.ValidateTokenFromOTP(context.Background(), test.token)

			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, test.wantResult, claims)
		})
	}
}

func TestValidateTokenFromFeature2FA(pt *testing.T) {
	newJwt, _ := New(
		&config.Config{}, &config.Secret{JWTSignatureKey: config.JWTSignatureKey{TokenOTPFeatureKey: "test-jwt-key-002"}}, nil,
	)

	tests := []struct {
		name       string
		token      string
		wantErr    string
		wantResult *otpModel.TokenOTPClaims
	}{
		{
			name:    "Invalid token key",
			token:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.DMp1LCRvKrG76ZKY99O7N0rrwaEsfbF-hm-jzKNVisY",
			wantErr: "token signature is invalid",
		},
		{
			name:  "Success",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1dWlkIjoiY2U5Mzc1OGUtYjE5Yi00M2MyLTk3MmEtZjY0YmUxNmYwOGNmIiwiaWRlbnRpZmllciI6IjMxYjNkMDk5LTg1MDgtNGU5My1hODBjLTA0NzFlZjIzN2E0YyJ9.VIf4_AUdXl-FM-CQtmm6pqY8MBkKdPLzmKU5UlC5zgc",
			wantResult: &otpModel.TokenOTPClaims{
				UUID:       "ce93758e-b19b-43c2-972a-f64be16f08cf",
				Identifier: "31b3d099-8508-4e93-a80c-0471ef237a4c",
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {
			claims, err := newJwt.ValidateTokenFromFeature2FA(context.Background(), test.token)

			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, test.wantResult, claims)
		})
	}
}

func TestGetIterateTokenFromRedis(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	scanKeyFormat := func(email, deviceId string) string {
		return "backend-portal:access-token:" + email + ":" + deviceId
	}
	pkgJwt, _ := New(&config.Config{}, &config.Secret{}, redisExt.WrapRedisClient(db, nil))

	validDeviceId := "test-device"

	tests := []struct {
		name       string
		email      string
		deviceId   string
		mockSetup  func(email, deviceId string, mock redismock.ClientMock)
		wantErr    string
		wantResult string
	}{
		{
			name: "ERROR: Token session",
			mockSetup: func(email, deviceId string, mock redismock.ClientMock) {
				mock.ExpectGet(scanKeyFormat(email, deviceId)).SetErr(errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name:     "ERROR: Redis nil result",
			email:    "email@example.id",
			deviceId: validDeviceId,
			mockSetup: func(email, deviceId string, mock redismock.ClientMock) {
				mock.ExpectGet(scanKeyFormat(email, deviceId)).RedisNil()
			},
			wantErr: "redis: nil",
		},
		{
			name:     "SUCCESS",
			email:    "hero@example.id",
			deviceId: validDeviceId,
			mockSetup: func(email, deviceId string, mock redismock.ClientMock) {
				mock.ExpectGet(scanKeyFormat(email, deviceId)).SetVal("valid-token")
			},
			wantResult: "valid-token",
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(test.email, test.deviceId, clientMock)

			token, err := pkgJwt.GetTokenLoggedInDevices(context.Background(), test.email, test.deviceId)
			if test.wantErr != "" {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
				return
			}
			require.Nil(t, err)
			assert.Equal(t, test.wantResult, token)
		})
	}
}

func TestRemoveIterateTokenFromRedis(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	scanKeyFormat := func(val string) string {
		return "backend-portal:access-token:" + val + ":*"
	}
	pkgJwt, _ := New(&config.Config{}, &config.Secret{}, redisExt.WrapRedisClient(db, nil))

	tests := []struct {
		name      string
		email     string
		mockSetup func(email string, mock redismock.ClientMock)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetErr(errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name:  "ERROR:Key not found",
			email: "invalid@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"INVALID-KEY"}, 0)
				mock.ExpectDel("INVALID-KEY").RedisNil()
			},
			wantErr: "redis: nil",
		},
		{
			name:  "SUCCESS",
			email: "hero@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"VALID-KEY"}, 0)
				mock.ExpectDel("VALID-KEY").SetVal(0)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(test.email, clientMock)
			if err := pkgJwt.RemoveIterateTokenFromRedis(context.Background(), test.email); test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestTerminateTokenOtherDevices(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	scanKeyFormat := func(val string) string {
		return "backend-portal:access-token:" + val + ":*"
	}
	pkgJwt, _ := New(&config.Config{}, &config.Secret{}, redisExt.WrapRedisClient(db, nil))

	tests := []struct {
		name      string
		email     string
		mockSetup func(email string, mock redismock.ClientMock)
		wantErr   string
	}{
		{
			name: "ERROR:Invalid session",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetErr(errors.New("invalid session"))
			},
			wantErr: "invalid session",
		},
		{
			name:  "ERROR:Key not found",
			email: "invalid@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"INVALID-KEY"}, 0)
				mock.ExpectSet("INVALID-KEY", constant.UserLoginSessionTerminated, 5*time.Minute).RedisNil()
			},
			wantErr: "redis: nil",
		},
		{
			name:  "SUCCESS",
			email: "hero@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"VALID-KEY"}, 0)
				mock.ExpectSet("VALID-KEY", constant.UserLoginSessionTerminated, 5*time.Minute).SetVal("")
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(test.email, clientMock)
			if err := pkgJwt.TerminateTokenOtherDevices(context.Background(), test.email, "device"); test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestTerminateTokenOfUserRoleChanged(pt *testing.T) {
	db, clientMock := redismock.NewClientMock()

	scanKeyFormat := func(val string) string {
		return "backend-portal:access-token:" + val + ":*"
	}
	pkgJwt, _ := New(&config.Config{}, &config.Secret{}, redisExt.WrapRedisClient(db, nil))

	tests := []struct {
		name      string
		email     string
		mockSetup func(email string, mock redismock.ClientMock)
		wantErr   string
	}{
		{
			name: "ERROR: Scan returns error",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetErr(errors.New("redis scan error"))
			},
			wantErr: "redis scan error",
		},
		{
			name:  "ERROR: Set returns error on first key",
			email: "user@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"KEY-1", "KEY-2"}, 0)
				mock.ExpectSet("KEY-1", constant.UserLoginRoleSessionChanged, 5*time.Minute).RedisNil()
			},
			wantErr: "redis: nil",
		},
		{
			name:  "ERROR: Set returns error on second key",
			email: "user2@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"KEY-1", "KEY-2"}, 0)
				mock.ExpectSet("KEY-1", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetVal("")
				mock.ExpectSet("KEY-2", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetErr(errors.New("set failed"))
			},
			wantErr: "set failed",
		},
		{
			name:  "SUCCESS: Single key",
			email: "single@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"SINGLE-KEY"}, 0)
				mock.ExpectSet("SINGLE-KEY", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetVal("")
			},
		},
		{
			name:  "SUCCESS: Multiple keys",
			email: "multi@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{"KEY-1", "KEY-2", "KEY-3"}, 0)
				mock.ExpectSet("KEY-1", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetVal("")
				mock.ExpectSet("KEY-2", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetVal("")
				mock.ExpectSet("KEY-3", constant.UserLoginRoleSessionChanged, 5*time.Minute).SetVal("")
			},
		},
		{
			name:  "SUCCESS: No keys found",
			email: "nokeys@example.id",
			mockSetup: func(email string, mock redismock.ClientMock) {
				mock.ExpectScan(0, scanKeyFormat(email), 0).SetVal([]string{}, 0)
			},
		},
	}
	for _, test := range tests {
		pt.Run(test.name, func(t *testing.T) {

			test.mockSetup(test.email, clientMock)
			if err := pkgJwt.TerminateTokenOfUserRoleChanged(context.Background(), test.email); test.wantErr == "" {
				require.Nil(t, err)

			} else {
				require.NotNil(t, err)
				assert.ErrorContains(t, err, test.wantErr)
			}
		})
	}
}

func TestNewJWT(t *testing.T) {
	pwd, _ := os.Getwd()
	projectRoot, _ := util.FindProjectRoot(pwd, "backend-portal")
	dir := path.Join(projectRoot, "test", "dummies")

	tests := []struct {
		name      string
		config    *config.Config
		secret    *config.Secret
		wantError error
	}{
		{
			name:      "SUCCESS: Without merchant rsa key pair",
			secret:    &config.Secret{},
			wantError: nil,
		},
		{
			name: "ERROR: Private key file not found",
			secret: &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantRSAKey: &config.RSAKeyPairFile{
						PrivateKeyFile: "./file-not-found",
					},
				},
			},
			wantError: errors.New("failed to read private key file: open ./file-not-found: no such file or directory"),
		},
		{
			name: "ERROR: Invalid private key value",
			secret: &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantRSAKey: &config.RSAKeyPairFile{
						PrivateKeyFile: dir + "/invalid_id_rsa",
					},
				},
			},
			wantError: errors.New("failed to parse private key from PEM: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key"),
		},
		{
			name: "ERROR: Invalid private key value",
			secret: &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantRSAKey: &config.RSAKeyPairFile{
						PrivateKeyFile: dir + "/valid_id_rsa",
						PublicKeyFile:  "./file-not-found",
					},
				},
			},
			wantError: errors.New("failed to read public key file: open ./file-not-found: no such file or directory"),
		},
		{
			name: "ERROR: Invalid private key value",
			secret: &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantRSAKey: &config.RSAKeyPairFile{
						PrivateKeyFile: dir + "/valid_id_rsa",
						PublicKeyFile:  dir + "/invalid_id_rsa.pub",
					},
				},
			},
			wantError: errors.New("failed to parse public key from PEM: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key"),
		},
		{
			name: "SUCCESS: With merchant RSA key pair",
			secret: &config.Secret{
				JWTSignatureKey: config.JWTSignatureKey{
					MerchantRSAKey: &config.RSAKeyPairFile{
						PrivateKeyFile: dir + "/valid_id_rsa",
						PublicKeyFile:  dir + "/valid_id_rsa.pub",
					},
				},
			},
			wantError: nil,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(test.config, test.secret, nil)
			assert.Equal(t, test.wantError, err)
		})
	}
}
