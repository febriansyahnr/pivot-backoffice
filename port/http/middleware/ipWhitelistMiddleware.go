package middleware

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func IPWhitelistMiddleware(svc service.IIPWhitelistService, headerKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, segment := otelTracer.Start(r.Context(), "port/http/middleware/IPWhitelistMiddleware")
			defer segment.End()

			var merchantId string

			if headerKey == constant.HeaderAuthorization {
				merchantClaims, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
				if !ok || merchantClaims == nil {
					response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
					return
				}
				merchantId = merchantClaims.MerchantId
			} else {
				merchantId = r.Header.Get(headerKey)
			}

			ip := r.Header.Get(constant.HeaderXRealIP)

			// When the ip is not set
			// we allow the request to proceed due internal services requests
			if ip == "" {
				next.ServeHTTP(w, r)
				return
			}

			if err := svc.ValidateIP(ctx, merchantId, ip); err != nil {
				response.SendOpenApiResponseError(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
