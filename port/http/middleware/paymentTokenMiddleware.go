package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"go.uber.org/zap"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func PaymentTokenMiddleware(jwtCore jwt.IJwt, redis redisExt.IRedisExt, logger logger.ILogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := r.Context()

			err := func() error {
				ctx, segment := otelTracer.Start(ctxReq, "http/middleware/PaymentTokenMiddleware")
				defer segment.End()

				// Get token from query params
				tokenString := r.URL.Query().Get("token")
				if tokenString == "" {
					tokenString = r.Header.Get(constant.HeaderXPaymentToken)
				}
				if tokenString == "" {
					logger.Warn(ctx, "[PaymentTokenMiddleware] empty request payment token")
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				// Parse token
				claims, errVerif := jwtCore.ValidatePaymentToken(tokenString)
				if errVerif != nil {
					logger.Warn(ctx, "[PaymentTokenMiddleware] invalid claims token", zap.Error(errVerif))
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				// check if token is expired
				// if not expired, return nil
				if time.Now().UTC().Unix() > claims.ExpiresAt.Unix() {
					logger.Warn(ctx, "[PaymentTokenMiddleware] token is expired")
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrTokenIsExpired)
				}

				// Check redis
				hashedToken := util.HashString(tokenString)
				validToken := false
				if err := redis.Get(ctx, fmt.Sprintf(constant.PaymentTokenCacheKey, hashedToken)).Scan(&validToken); err != nil || !validToken {
					logger.Warn(ctx, "[PaymentTokenMiddleware] invalid get token in redis", zap.Error(err))
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				// Add payment id to context
				ctxReq = context.WithValue(ctxReq, constant.CtxPaymentID, claims.UUID)

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
