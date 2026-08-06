package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
)

type MiddlewareFunc func(next http.Handler) http.Handler

func AuthMiddleware(
	jwtCore jwt.IJwt,
	redis redisExt.IRedisExt) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := r.Context()
			err := func() error {
				ctx, segment := otelTracer.Start(ctxReq, "http/middleware/AuthMiddleware")
				defer segment.End()

				// Get token from authorization header
				tokenHeader := r.Header.Get(constant.HeaderAuthorization)

				if !strings.Contains(tokenHeader, "Bearer") {
					return pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("invalid token"))
				}

				tokenString := strings.Replace(tokenHeader, "Bearer ", "", -1)

				// Parse token
				claims, errVerif := jwtCore.Verify(ctx, tokenString)
				if errVerif != nil {
					return pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("invalid token"))
				}

				// check if token is expired
				// if not expired, return nil
				if time.Now().UTC().Unix() > claims.ExpiresAt.Unix() {
					return pkgErrors.New(response.HttpErrUnauthorized, fmt.Errorf("token is expired"))
				}

				// Add user info to context
				ctxReq = context.WithValue(ctxReq, constant.CtxUserInfoKey, claims)
				//r = r.WithContext(ctx)

				// Add merchant id to context
				ctxReq = context.WithValue(ctxReq, constant.CtxMerchantIDKey, claims.MerchantId)

				// Logging Purpose
				ctxReq = context.WithValue(ctxReq, pdkConstant.CtxMerchantIDKey, claims.MerchantId)
				ctxReq = context.WithValue(ctxReq, pdkConstant.CtxUserIDKey, claims.UUID)
				ctxReq = context.WithValue(ctxReq, pdkConstant.CtxUserRoleKey, claims.Role)

				// Check token in redis
				if _, err := jwtCore.GetTokenLoggedInDevices(ctx, claims.Email, claims.DeviceIdentifier); err != nil {
					if errors.Is(err, constant.ErrLoginSessionTerminated) {
						return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrLoginSessionTerminated)
					}

					if errors.Is(err, constant.ErrLoginRoleSessionChanged) {
						return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrLoginRoleSessionChanged)
					}

					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				return nil
			}()

			if err != nil {
				response.SendApiResponseError(ctxReq, w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctxReq))
		})
	}
}
