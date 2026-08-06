package notification_test

import (
	"context"
	"errors"
	"fmt"
	pdkLoggerMock "github.com/paper-indonesia/pivot-backoffice/mocks/pdk/logger"
	loggerMock "github.com/paper-indonesia/pdk/v2/logger"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	rabbitMQExtMock "github.com/paper-indonesia/pivot-backoffice/mocks/pkg/rabbitmqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/consumer/notification"

	"github.com/paper-indonesia/pdk/v2/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryNotification(t *testing.T) {
	logger := pdkLoggerMock.NewILogger(t)
	rmq := rabbitMQExtMock.NewRabbitMQExt(t)

	handler := New(logger, rmq)

	tests := []struct {
		name      string
		message   *amqp.Delivery
		setupMock func()
		wantErr   error
	}{
		{
			name: "ERROR:Retry limit exceeded",
			message: &amqp.Delivery{
				Headers: amqp.Table{
					rabbitMqExt.HeaderXRetryCount: 6,
				},
			},
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Retry Notification", loggerMock.Any("details", map[string]interface{}{
						"status": "FAILED", "message": "retry limit exceeded", "routingKey": "", "retryCount": int64(6), "payload": "",
					}),
				).Once().Return()
			},
			wantErr: errors.New("retry limit exceeded"),
		},
		{
			name:    "ERROR:Empty routing key",
			message: &amqp.Delivery{},
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Retry Notification", loggerMock.Any("details", map[string]interface{}{
						"status": "FAILED", "message": "empty routing key", "routingKey": "", "retryCount": int64(0), "payload": "",
					}),
				).Once().Return()
			},
			wantErr: errors.New("empty routing key"),
		},
		{
			name: "ERROR:Invalid body request",
			message: &amqp.Delivery{
				RoutingKey: "key",
				Body:       []byte(`1`),
			},
			setupMock: func() {
				logger.On(
					"Info", mock.Anything, "Retry Notification", loggerMock.Any("details", map[string]interface{}{
						"status": "FAILED", "message": "Unmarshal JSON: json: cannot unmarshal number into Go value of type notification.PushNotificationPayload", "routingKey": "key", "retryCount": int64(0), "payload": "1",
					}),
				).Once().Return()
			},
			wantErr: errors.New("Unmarshal JSON: json: cannot unmarshal number into Go value of type notification.PushNotificationPayload"),
		},
		{
			name: "ERROR:Some error",
			message: &amqp.Delivery{
				RoutingKey: "abc",
				Body:       []byte(`{}`),
			},
			setupMock: func() {
				rmq.On(
					"PushNotification", mock.Anything, mock.Anything,
				).Once().Return(constant.ErrSomeErrorForUnitTest)
				logger.On(
					"Info", mock.Anything, "Retry Notification", loggerMock.Any("details", map[string]interface{}{
						"status": "FAILED", "message": "Push Notification: some error", "routingKey": "abc", "retryCount": int64(0), "payload": "{}",
					}),
				).Once().Return()
			},
			wantErr: fmt.Errorf("Push Notification: %w", constant.ErrSomeErrorForUnitTest),
		},
		{
			name: "SUCCESS",
			message: &amqp.Delivery{
				RoutingKey: "key1",
				Body:       []byte(`{}`),
			},
			setupMock: func() {
				rmq.On(
					"PushNotification", mock.Anything, mock.Anything,
				).Return(nil)
				logger.On(
					"Info", mock.Anything, "Retry Notification", loggerMock.Any("details", map[string]interface{}{
						"status": "SUCCESS", "message": "successfully retry notification", "routingKey": "key1", "retryCount": int64(0), "payload": "{}",
					}),
				).Once().Return()
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			assert.Equal(t, test.wantErr, handler.RetryNotification(context.Background(), test.message))
		})
	}
}
