package jwt

import (
	"context"
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/merchant"
	otpModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/otp"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/payment"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/user"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"

	goJwt "github.com/golang-jwt/jwt/v5"
)

type IJwt interface {
	GenerateAccessToken(ctx context.Context, user *userModel.User) (token string, err error)
	GenerateRefreshToken(ctx context.Context, user *userModel.User, expiredAt time.Time) (token string, err error)
	Verify(ctx context.Context, tokenString string) (claims *userModel.UserTokenClaims, err error)
	VerifyMerchantToken(ctx context.Context, tokenString string) (claims *merchant.MerchantAuthTokenClaims, err error)
	GenerateMerchantToken(ctx context.Context, clientID, merchantID string) (token string, err error)
	GetTokenLoggedInDevices(ctx context.Context, email, deviceId string) (token string, err error)
	RemoveIterateTokenFromRedis(ctx context.Context, email string) (err error)
	TerminateTokenOtherDevices(ctx context.Context, email, deviceId string) (err error)
	TerminateTokenOfUserRoleChanged(ctx context.Context, email string) (err error)

	ValidateTokenFromOTP(ctx context.Context, tokenString string) (claims *otpModel.TokenOTPClaims, err error)
	ValidateTokenFromFeature2FA(ctx context.Context, tokenString string) (claims *otpModel.TokenOTPClaims, err error)

	GeneratePaymentToken(paymentID string, expiredAt time.Time) (token string, err error)
	ValidatePaymentToken(tokenString string) (claims *paymentModel.PaymentClaims, err error)
}

type jwtConfig struct {
	Config *config.Config
	Secret *config.Secret
	Redis  redisExt.IRedisExt

	// Internal Used
	merchantRSAKeyPair *merchantRSAKeyPair
}

type merchantRSAKeyPair struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

const accessTokenKeyFormat = "backend-portal:access-token:%s:%s"

func New(config *config.Config, secret *config.Secret, redis redisExt.IRedisExt) (IJwt, error) {
	jwt := &jwtConfig{
		Config: config,
		Secret: secret,
		Redis:  redis,
	}
	if secret.JWTSignatureKey.MerchantRSAKey != nil {
		privateKeyPEM, err := os.ReadFile(secret.JWTSignatureKey.MerchantRSAKey.PrivateKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key file: %v", err)
		}
		privateKey, err := goJwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key from PEM: %v", err)
		}

		publicKeyPEM, err := os.ReadFile(secret.JWTSignatureKey.MerchantRSAKey.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file: %v", err)
		}
		publicKey, err := goJwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key from PEM: %v", err)
		}

		jwt.merchantRSAKeyPair = &merchantRSAKeyPair{privateKey, publicKey}
	}

	return jwt, nil
}

func (c *jwtConfig) GenerateAccessToken(ctx context.Context, user *userModel.User) (token string, err error) {
	userToken := userModel.UserTokenClaims{
		UUID:             user.UUID,
		Name:             user.Name,
		Email:            user.Email,
		Blocked:          user.Blocked.Time,
		MerchantId:       user.MerchantId,
		Role:             user.Role.String,
		RoleID:           user.RoleId.String,
		DeviceIdentifier: user.DeviceIdentifier,
	}

	userToken.RegisteredClaims = goJwt.RegisteredClaims{
		Subject:   user.UUID,
		Issuer:    c.Config.ServiceName,
		ExpiresAt: goJwt.NewNumericDate(time.Now().UTC().Add(constant.LOGIN_EXPIRATION_DURATION)),
	}

	newToken := goJwt.NewWithClaims(goJwt.SigningMethodHS256, userToken)

	// because the function only returns error when the signature key is not []byte, but
	// we always set the signature key as []byte,
	// so we can ignore the error
	return newToken.SignedString([]byte(c.Secret.JWTSignatureKey.UserKey))
}

func (c *jwtConfig) GenerateRefreshToken(
	ctx context.Context, user *userModel.User, expiredAt time.Time) (token string, err error) {

	userClaims := goJwt.RegisteredClaims{
		Subject:   user.UUID,
		Issuer:    c.Config.ServiceName,
		ExpiresAt: goJwt.NewNumericDate(expiredAt),
	}

	newToken := goJwt.NewWithClaims(goJwt.SigningMethodHS256, userClaims)

	// because the function only returns error when the signature key is not []byte, but
	// we always set the signature key as []byte,
	// so we can ignore the error
	return newToken.SignedString([]byte(c.Secret.JWTSignatureKey.UserKey))
}

func (c *jwtConfig) generateTokenOTP(_ context.Context, key string, data *otpModel.TokenOTPClaims) (string, error) {
	data.Issuer = c.Config.ServiceName

	return goJwt.NewWithClaims(goJwt.SigningMethodHS256, data).
		SignedString([]byte(key))
}

func (c *jwtConfig) Verify(ctx context.Context, tokenString string) (claims *userModel.UserTokenClaims, err error) {
	token, err := goJwt.ParseWithClaims(tokenString, &userModel.UserTokenClaims{},
		func(token *goJwt.Token) (interface{}, error) {
			return []byte(c.Secret.JWTSignatureKey.UserKey), nil
		})
	if err != nil {
		return nil, err
	}

	return token.Claims.(*userModel.UserTokenClaims), nil
}

func (c *jwtConfig) ValidateTokenFromOTP(ctx context.Context, tokenString string) (*otpModel.TokenOTPClaims, error) {
	return c.validateTokenOTP(ctx, c.Secret.JWTSignatureKey.TokenOTPKey, tokenString)
}

func (c *jwtConfig) ValidateTokenFromFeature2FA(ctx context.Context, tokenString string) (*otpModel.TokenOTPClaims, error) {
	return c.validateTokenOTP(ctx, c.Secret.JWTSignatureKey.TokenOTPFeatureKey, tokenString)
}

func (c *jwtConfig) validateTokenOTP(_ context.Context, key, tokenString string) (*otpModel.TokenOTPClaims, error) {
	token, err := goJwt.ParseWithClaims(tokenString, &otpModel.TokenOTPClaims{},
		func(token *goJwt.Token) (interface{}, error) {
			return []byte(key), nil
		})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*otpModel.TokenOTPClaims), nil
}

// GetIterateTokenFromRedis is a function that iterate all deviceID to get certain token
func (c *jwtConfig) GetTokenLoggedInDevices(ctx context.Context, email, deviceId string) (token string, err error) {
	accessToken, errGet := c.Redis.Get(ctx, fmt.Sprintf(accessTokenKeyFormat, email, deviceId)).Result()
	if errGet != nil {
		return "", errGet
	}

	if accessToken == constant.UserLoginSessionTerminated {
		return "", constant.ErrLoginSessionTerminated
	}

	if accessToken == constant.UserLoginRoleSessionChanged {
		return "", constant.ErrLoginRoleSessionChanged
	}

	return accessToken, nil
}

// RemoveIterateTokenFromRedis is a function that iterate all deviceID to remove certain token
func (c *jwtConfig) RemoveIterateTokenFromRedis(ctx context.Context, email string) (err error) {
	// Check token in redis
	keysIterator := c.Redis.
		Scan(ctx, 0, fmt.Sprintf(accessTokenKeyFormat, email, "*"), 0).Iterator()

	for keysIterator.Next(ctx) {
		if err = c.Redis.Del(ctx, keysIterator.Val()).Err(); err != nil {
			return
		}
	}
	return keysIterator.Err()
}

// TerminateTokenOtherDevices is a function that iterate all deviceID to terminated token other device
func (c *jwtConfig) TerminateTokenOtherDevices(ctx context.Context, email, deviceId string) (err error) {

	// Check token in redis
	keysIterator := c.Redis.
		Scan(ctx, 0, fmt.Sprintf(accessTokenKeyFormat, email, "*"), 0).Iterator()

	for keysIterator.Next(ctx) {
		key := keysIterator.Val()
		if key == fmt.Sprintf(accessTokenKeyFormat, email, deviceId) {
			continue
		}

		_, err = c.Redis.Set(ctx, key, constant.UserLoginSessionTerminated, 5*time.Minute).Result()
		if err != nil {
			return
		}
	}
	return keysIterator.Err()
}

func (c *jwtConfig) TerminateTokenOfUserRoleChanged(ctx context.Context, email string) (err error) {

	// Check token in redis
	keysIterator := c.Redis.
		Scan(ctx, 0, fmt.Sprintf(accessTokenKeyFormat, email, "*"), 0).Iterator()

	for keysIterator.Next(ctx) {
		key := keysIterator.Val()

		_, err = c.Redis.Set(ctx, key, constant.UserLoginRoleSessionChanged, 5*time.Minute).Result()
		if err != nil {
			return
		}
	}
	return keysIterator.Err()
}

func (c *jwtConfig) GeneratePaymentToken(paymentID string, expiredAt time.Time) (string, error) {
	data := &paymentModel.PaymentClaims{
		UUID: paymentID,
	}
	data.Issuer = c.Config.ServiceName
	data.ExpiresAt = goJwt.NewNumericDate(expiredAt)

	return goJwt.NewWithClaims(goJwt.SigningMethodHS256, data).
		SignedString([]byte(c.Secret.JWTSignatureKey.PaymentToken))
}

func (c *jwtConfig) ValidatePaymentToken(tokenString string) (*paymentModel.PaymentClaims, error) {
	token, err := goJwt.ParseWithClaims(tokenString, &paymentModel.PaymentClaims{},
		func(token *goJwt.Token) (interface{}, error) {
			return []byte(c.Secret.JWTSignatureKey.PaymentToken), nil
		})
	if err != nil {
		return nil, err
	}
	return token.Claims.(*paymentModel.PaymentClaims), nil
}
