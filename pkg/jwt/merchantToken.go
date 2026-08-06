package jwt

import (
	"context"
	"errors"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	goJwt "github.com/golang-jwt/jwt/v5"
)

func (c *jwtConfig) VerifyMerchantToken(_ context.Context, tokenString string) (claims *merchant.MerchantAuthTokenClaims, err error) {
	token, err := goJwt.ParseWithClaims(tokenString, &merchant.MerchantAuthTokenClaims{}, func(t *goJwt.Token) (any, error) {
		if _, ok := t.Method.(*goJwt.SigningMethodHMAC); ok {
			return []byte(c.Secret.JWTSignatureKey.MerchantKey), nil

		} else if _, ok := t.Method.(*goJwt.SigningMethodRSA); ok && c.merchantRSAKeyPair != nil {
			return c.merchantRSAKeyPair.publicKey, nil
		}
		return nil, errors.New("invalid method")
	})
	if err != nil {
		if errors.Is(err, goJwt.ErrTokenExpired) {
			return nil, constant.ErrExpiredMerchantAuth
		}
		return nil, err
	}
	return token.Claims.(*merchant.MerchantAuthTokenClaims), nil
}

func (c *jwtConfig) GenerateMerchantToken(_ context.Context, clientID, merchantID string) (token string, err error) {

	if c.merchantRSAKeyPair == nil {
		return "", errors.New("requires merchant rsa key pair")
	}

	merchantToken := merchant.MerchantAuthTokenClaims{
		ClientId:   clientID,
		MerchantId: merchantID,
	}

	merchantToken.RegisteredClaims = goJwt.RegisteredClaims{
		Subject:   clientID,
		Issuer:    c.Config.ServiceName,
		ExpiresAt: goJwt.NewNumericDate(time.Now().UTC().Add(constant.MerchantAuthExpirationDuration)),
	}

	return goJwt.NewWithClaims(goJwt.SigningMethodRS256, merchantToken).
		SignedString(c.merchantRSAKeyPair.privateKey)
}
