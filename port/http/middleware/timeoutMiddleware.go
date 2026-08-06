package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// DynamicTimeout creates a middleware that applies different timeout durations based on request path
// It accepts a map of path patterns to timeout durations (in seconds)
// Any path not matched in the map will use the defaultTimeout
func DynamicTimeout(log logger.ILogger, defaultDuration int) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var event = ""

			switch r.URL.Path {
			case "/api/v1/disbursements/approval-actions",
				"/api/v1/disbursements/single/retry",
				"/open-api/v1/inquiry-account",
				"/open-api/v1/payouts",
				"/open-api/v1/payments",
				"/internal/v1/payouts",
				"/internal/v1/payments":
				event = constant.TimeoutEventSnapCore
			case "/open-api/v2/payments":
				event = constant.TimeoutEventUnifiedPayment
			case "/api/v1/disbursements/bulk/validate",
				"/api/v1/disbursements/bulk/preview",
				"/api/v1/disbursements/bulk/upload":
				event = constant.TimeoutEventBulkDisbursement
			}

			if event != "" {
				ffContext := ffcontext.NewEvaluationContext(uuid.NewString())
				ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameEvent, event)
				defaultDuration, _ = ffclient.IntVariation(constant.FeatureFlagKeyCustomContextTimeout, ffContext, defaultDuration)

				log.Info(r.Context(), "dynamic timeout applied",
					logger.Int("duration", defaultDuration),
					logger.String("path", r.URL.Path),
				)
			}

			// Create timeout context
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(defaultDuration)*time.Second)
			defer func() {
				cancel()
				if ctx.Err() == context.DeadlineExceeded {
					deadlineTime, _ := ctx.Deadline()
					log.Info(ctx, "context deadline timeout", logger.String("timeout", deadlineTime.Format(time.RFC3339Nano)), logger.Int("defaultDurationInSeconds", defaultDuration))
					w.WriteHeader(http.StatusGatewayTimeout)
				}
			}()

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
