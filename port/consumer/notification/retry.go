package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (h *handler) RetryNotification(ctx context.Context, msg *rabbitMqExt.Delivery) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/notification/RetryNotification")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			h.logger.Error(ctx, "Panic recovery from RetryNotification", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var retryCount int64
	if rt, ok := msg.Headers[rabbitMqExt.HeaderXRetryCount]; ok {
		retryCount, _ = strconv.ParseInt(fmt.Sprintf("%v", rt), 10, 64)
	}

	ctx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	payload := notification.PushNotificationPayload{}

	defer func() {
		details := map[string]interface{}{
			"status":     "SUCCESS",
			"message":    "successfully retry notification",
			"routingKey": msg.RoutingKey,
			"retryCount": retryCount,
			"payload":    string(msg.Body),
		}
		if err != nil {
			details["status"], details["message"] = "FAILED", err.Error()
		}
		h.logger.Info(ctx, "Retry Notification", logger.Any("details", details))
	}()

	if retryCount > constant.NotificationRetryLimit {
		return errors.New("retry limit exceeded")

	} else if msg.RoutingKey == "" {
		return errors.New("empty routing key")
	}

	if err = json.Unmarshal(msg.Body, &payload); err != nil {
		return fmt.Errorf("Unmarshal JSON: %v", err)
	}

	err = h.rmq.PushNotification(
		ctx, &notification.PushNotification{RoutingKey: msg.RoutingKey, Payload: payload, RetryCount: int(retryCount) + 1},
	)
	if err != nil {
		return fmt.Errorf("Push Notification: %w", err)
	}
	return nil
}
