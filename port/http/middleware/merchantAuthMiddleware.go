package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/jwt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
)

func MerchantAuthMiddleware(jwtCore jwt.IJwt) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := r.Context()

			err := func() error {
				ctx, segment := otelTracer.Start(ctxReq, "http/middleware/MerchantAuthMiddleware")
				defer segment.End()

				// Get token from authorization header
				tokenHeader := r.Header.Get("Authorization")
				if !strings.Contains(tokenHeader, "Bearer") {
					return errors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				tokenString := strings.Replace(tokenHeader, "Bearer ", "", -1)

				// Parse token
				claims, errVerif := jwtCore.VerifyMerchantToken(ctx, tokenString)
				if errVerif != nil {
					return errors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				// check if token is expired
				// if not expired, return nil
				if time.Now().UTC().Unix() > claims.ExpiresAt.Unix() {
					return errors.New(response.HttpErrUnauthorized, fmt.Errorf("token is expired"))
				}

				// Add user info to context
				ctxReq = context.WithValue(ctxReq, constant.CtxMerchantInfo, claims)

				// Add merchant id to context
				ctxReq = context.WithValue(ctxReq, constant.CtxMerchantIDKey, claims.MerchantId)
				ctxReq = context.WithValue(ctxReq, pdkConstant.CtxMerchantIDKey, claims.MerchantId)

				return nil
			}()

			if err != nil {
				response.SendOpenApiNonSnapResponseError(ctxReq, w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctxReq))
		})
	}
}
