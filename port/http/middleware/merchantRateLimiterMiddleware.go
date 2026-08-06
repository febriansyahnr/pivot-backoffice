package middleware

import (
	"net"
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	ratelimiter "github.com/paper-indonesia/pivot-backoffice/internal/model/rateLimiter"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func MerchantRateLimiterMiddleware(service service.IRateLimiter, cfg *config.Config) MiddlewareFunc {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, segment := otelTracer.Start(r.Context(), "port/http/middleware/MerchantRateLimiterMiddleware")
			defer segment.End()

			ffContext := ffcontext.NewEvaluationContext(cfg.Environment)
			ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEnv, cfg.Environment)
			enabled, _ := ffclient.BoolVariation(constant.FeatureFlagKeyEnableMerchantRateLimitMiddleware, ffContext, false)

			if !enabled {
				next.ServeHTTP(w, r)
				return
			}

			merchantClaims, ok := r.Context().Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
			if !ok || merchantClaims == nil {
				response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
				return
			}

			path := r.URL.Path
			ip := r.Header.Get(constant.HeaderXRealIP)
			if net.ParseIP(ip) == nil {
				ip = ""
			}

			metadata, err := service.ValidateMerchantRateLimit(r.Context(), ratelimiter.MerchantRateLimitRequest{
				MerchantID: merchantClaims.MerchantId,
				Path:       path,
				IPAddress:  ip,
				HTTPMethod: r.Method,
			})

			if metadata != nil && metadata.RateLimitLimit > 0 {
				w.Header().Set("Rate-Limit-Limit", strconv.Itoa(metadata.RateLimitLimit))
				w.Header().Set("Rate-Limit-Remaining", strconv.Itoa(metadata.RateLimitRemaining))
				w.Header().Set("Rate-Limit-Reset", strconv.FormatInt(metadata.RateLimitReset, 10))
			}

			if err != nil {
				response.SendOpenApiResponseError(w, err)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
