package middleware

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	ffcontext "github.com/thomaspoignant/go-feature-flag/ffcontext"
)

func HealthCheckContextMiddleware(logger pdkLogger.ILogger, ctx context.Context, dependentService string, dependentHealthCheckFunc func(ctx context.Context) error) error {
	var startTime = time.Now()
	ffContext := ffcontext.NewEvaluationContext(uuid.NewString())
	ffContext.AddCustomAttribute(constant.FeatureFlagTargetQueryNameDependentService, dependentService)
	healthCheckTimeout, err := ffclient.IntVariation(constant.FeatureFlagKeyHealthCheckCustomContextTimeout, ffContext, constant.HealthCheckDefaultTimeoutInMs)
	if err != nil {
		logger.Warn(context.Background(), "failed to get feature flag for health check custom context timeout", pdkLogger.Error(err), pdkLogger.String("dependentService", dependentService))
		healthCheckTimeout = constant.HealthCheckDefaultTimeoutInMs
	}

	traceId, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string)
	newCtx, cancel := context.WithTimeout(ctx, time.Duration(healthCheckTimeout)*time.Millisecond)
	newCtx = context.WithValue(newCtx, pdkConst.CtxTraceIdKey, traceId)
	defer func() {
		cancel()
		if err := newCtx.Err(); err != nil {
			if err == context.DeadlineExceeded {
				deadlineTime, _ := newCtx.Deadline()
				logger.Info(newCtx, "health check context deadline timeout", pdkLogger.String("dependentService", dependentService), pdkLogger.String("startTime", startTime.Format(time.RFC3339Nano)), pdkLogger.String("timeout", deadlineTime.Format(time.RFC3339Nano)), pdkLogger.Int("defaultDurationMs", healthCheckTimeout))
			}
		}
	}()

	return dependentHealthCheckFunc(newCtx)
}
