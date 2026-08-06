package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	chiExtMiddleware "github.com/paper-indonesia/pdk/v2/chiExt/middleware"
)

// ApprovalActionsIdempotencyMiddleware creates a middleware for handling idempotency
// specifically for approval actions with daily limit error cleanup
func ApprovalActionsIdempotencyMiddleware(redis redisExt.IRedisExt, serviceName string, log logger.ILogger) MiddlewareFunc {
	const headerIdempotencyKey = "X-Idempotent-Key"

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if wr, ok := w.(*chiExtMiddleware.ResponseWriter); ok {
				defer func() {
					hasErrInvalidPayoutItem := wr.Status == http.StatusUnprocessableEntity && (strings.Contains(wr.BodyString(), constant.ErrInvalidBatchPayoutItem.Error()))

					if wr.Status == http.StatusTooManyRequests && (strings.Contains(wr.BodyString(), constant.ErrDailyLimitReached.Error()) ||
						(strings.Contains(wr.BodyString(), response.HttpStatusErrorDailyLimitReached) && strings.Contains(wr.BodyString(), strings.Split(constant.ErrMsgPayoutDailyLimitRemainingToday, ":")[0]+":"))) || hasErrInvalidPayoutItem {
						idempotencyKey := fmt.Sprintf(
							"backend-portal:pdk-idempotency:POST:approval-actions:%s", r.Header.Get(headerIdempotencyKey),
						)
						if err := redis.Del(context.Background(), idempotencyKey).Err(); err != nil {
							log.Error(r.Context(), "error when delete idempotency key on approval actions", zap.Error(err))
						}
					}
				}()
			}
			IdempotencyApiMiddleware(redis, serviceName, "approval-actions", headerIdempotencyKey)(h).ServeHTTP(w, r)
		})
	}
}
