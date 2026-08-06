package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/idempotenshine"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	ffcontext "github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// CacheHTTPRequestMiddleware creates a middleware that caches HTTP requests based on idempotency keys.
// The middleware leverages the idempotenshine library to ensure that identical requests (identified by the same
// idempotency key) return consistent responses within a configured time period.
//
// When an idempotency key is present in the X-Idempotency-Key header and caching is enabled via feature flags,
// the request will be cached for the duration specified by the feature flag. Subsequent identical requests
// will return the cached result rather than executing the handler again.
func CacheHTTPRequestMiddleware(config *config.Config, log logger.ILogger, redisClient redisExt.IRedisExt) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				defaultDuration int
				ctx             = r.Context()
				idempotencyKey  = r.Header.Get(constant.HeaderXIdempotencyKey)
			)

			ffContext := ffcontext.NewEvaluationContext(uuid.NewString())
			ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameURLPath, r.URL.Path)
			defaultDuration, _ = ffclient.IntVariation(constant.FeatureFlagKeyHttpRequestCacheDuration, ffContext, defaultDuration)

			if defaultDuration > 0 && idempotencyKey != "" {
				keySource := idempotenshine.HeaderKeySource(constant.HeaderXIdempotencyKey)
				idempotenMiddleware := idempotenshine.Middleware(
					config.ServiceName,
					r.URL.Path,
					idempotenshine.WithTTL(time.Duration(defaultDuration)*time.Second),
					idempotenshine.WithKeySource(keySource),
					idempotenshine.WithRedisClient(redisClient.Client()),
					idempotenshine.WithLogicOptionUsage(idempotenshine.LogicOption{
						ReturnDataWhenKeyExists: true,
					}),
					idempotenshine.WithClientResponder(WithHttpCacheApiClientResponder(constant.HeaderXIdempotencyKey)),
				)

				log.Info(ctx, "using idempotency middleware", logger.String("idempotencyKey", idempotencyKey), logger.Int("duration", defaultDuration))

				idempotenMiddleware(next).ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// WithHttpCacheApiClientResponder creates a response handler function for the idempotenshine middleware
// that handles cancellation of API client requests.
// It takes a headerKey parameter that specifies which header to check for an idempotency key.
// The returned handler function will:
// - Return an error if the idempotency key header is missing
// - Handle in-progress request errors specifically
// - Fall back to a generic internal server error for other error cases
func WithHttpCacheApiClientResponder(headerKey string) idempotenshine.ResponseHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Header.Get(headerKey) == "" {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdempotencyKeyRequired))
			return
		}

		// TODO: update pdk to provide inprogress error
		if err == idempotenshine.ErrRequestInProgress {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrRequestInProgress))
			return
		}

		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrInternal, constant.ErrInternalServerForUser))
	}
}
