package rabbitMqExt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/notification"
	activitypb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/activity"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var directExchangeType = "direct"

func (r *rabbitMQExt) Publish(
	ctx context.Context,
	document string,
	exchangeType *string,
	message interface{},
	failedMsg ...amqp.Delivery,
) error {
	rabbitMqConfig := routingRabbitMqConfig(document)

	rmqExchangeType := "topic"
	if exchangeType != nil {
		rmqExchangeType = *exchangeType
	}

	ch, err := r.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	args := amqp.Table{
		"x-queue-type": amqp.QueueTypeQuorum, // NOSONAR
	}
	if rabbitMqConfig.DLQExchange != "" {

		args["x-dead-letter-exchange"] = rabbitMqConfig.DLQExchange   // NOSONAR
		args["x-dead-letter-routing-key"] = rabbitMqConfig.RoutingKey // NOSONAR
		if rabbitMqConfig.DLQRoutingKey != "" {
			args["x-dead-letter-routing-key"] = rabbitMqConfig.DLQRoutingKey // NOSONAR
		}
	}

	err = ch.ExchangeDeclare(
		rabbitMqConfig.Exchange, // name
		rmqExchangeType,         // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when exchange declare", pdkLogger.Error(err))
		return err
	}

	_, err = ch.QueueDeclare(
		rabbitMqConfig.QueueName, // Queue name
		true,                     // Durable (the queue will not survive server restarts)
		false,                    // Delete when unused
		false,                    // Exclusive (queue only accessible by the connection that declares it)
		false,                    // No-wait
		args,                     // Arguments
	)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when queue declare", pdkLogger.Error(err))
		return err
	}

	err = ch.QueueBind(
		rabbitMqConfig.QueueName,  // Queue name
		rabbitMqConfig.RoutingKey, // Routing key
		rabbitMqConfig.Exchange,   // Exchange
		false,
		nil,
	)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when queue bind", pdkLogger.Error(err))
		return err
	}

	var body []byte

	switch payload := message.(type) {
	case []byte:
		body = payload

	default:
		if body, err = json.Marshal(payload); err != nil {
			r.logger.Error(ctx, "error rabbitmq when marshaling message", pdkLogger.Error(err))
			return err
		}
	}

	messageId, _ := ctx.Value(constant.CtxMessageId).(string)
	if messageId == "" {
		messageId = uuid.NewString()
	}

	// Here you use the function to prepare the message for retry
	publishing := amqp.Publishing{
		ContentType: "text/plain",
		Body:        body,
		Headers:     amqp.Table{},
		Timestamp:   time.Now().UTC(),
		MessageId:   messageId,
	}
	if expiration, ok := ctx.Value(constant.CtxRabbitMQExpiration).(int64); ok && expiration > 0 {
		publishing.Expiration = fmt.Sprintf("%d", expiration)
	}

	if len(failedMsg) > 0 {
		publishing = r.incrementRetryCountAndPrepareMessage(failedMsg[0])
	}

	if traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string); traceID != "" {
		publishing.Headers[HeaderTraceId] = traceID
	}
	if parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string); parentMerchantId != "" {
		publishing.Headers[HeaderXParentMerchantId] = parentMerchantId
	}
	if derivedMerchantId, _ := ctx.Value(constant.CtxDerivedMerchantID).(string); derivedMerchantId != "" {
		publishing.Headers[HeaderXDerivedMerchantId] = derivedMerchantId
	}
	if merchantId, _ := ctx.Value(constant.CtxMerchantIDKey).(string); merchantId != "" {
		publishing.Headers[HeaderXMerchantId] = merchantId
	}

	err = ch.PublishWithContext(
		ctx,
		rabbitMqConfig.Exchange,   // Exchange (empty string means the default exchange)
		rabbitMqConfig.RoutingKey, // Routing key (queue name is used as the routing key here)
		false,                     // Mandatory
		false,                     // Immediate
		publishing,
	)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when publish", pdkLogger.Error(err))
		return err
	}

	return nil
}

func (r *rabbitMQExt) PublishActivity(
	ctx context.Context,
	merchantID, userID *string,
	tag, activity string,
	parameter interface{},
) error {
	// Don't publish activity on empty merchantID
	if merchantID == nil {
		return nil
	}

	message := &activitypb.Activity{
		Id:          uuid.NewString(),
		MerchantId:  *merchantID,
		UserId:      userID,
		Tag:         tag,
		Activity:    activity,
		ServiceName: activity,
		CreatedAt:   timestamppb.New(time.Now().UTC()),
		UpdatedAt:   timestamppb.New(time.Now().UTC()),
	}
	if parameter != nil {
		buf, err := json.Marshal(parameter)
		if err != nil {
			return err
		}
		message.Parameter, _ = anypb.New(wrapperspb.Bytes(buf))
	}

	rawMessage, _ := proto.Marshal(message)

	return r.Publish(ctx, ActivityInsertRoutingKey, &directExchangeType, rawMessage)
}

func (r *rabbitMQExt) PushNotification(ctx context.Context, data *notification.PushNotification) (err error) {
	ch, err := r.getChannel()
	if err != nil {
		return fmt.Errorf("create channel: %v", err)
	}
	defer func() {
		if e := ch.Close(); e != nil {
			err = fmt.Errorf("close channel: %v", err)
		}
	}()

	if err = r.notificationExchangeDeclareSetting(ch); err != nil {
		return err
	}

	body, _ := json.Marshal(data.Payload)
	err = ch.PublishWithContext(
		ctx,                  // Context
		NotificationExchange, // Exchange
		data.RoutingKey,      // Key
		false,                // Mandatory
		false,                // Immediate,
		amqp.Publishing{
			Body:        body,
			Timestamp:   time.Now(),
			Priority:    data.Priority,
			MessageId:   data.Payload.ID,
			ContentType: constant.MIMEApplicationJSON,
			Headers: amqp.Table{
				HeaderXRetryCount: data.RetryCount,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("push notification message: %v", err)
	}
	return
}

func (r *rabbitMQExt) notificationExchangeDeclareSetting(ch *amqp.Channel) (err error) {
	// Unroutable Notification Exchange
	err = ch.ExchangeDeclare(
		UnroutedNotificationExchange, // Name
		amqp.ExchangeTopic,           // Kind
		true,                         // Durable
		false,                        // Auto delete
		false,                        // Internal
		false,                        // No wait
		nil,                          // Arguments
	)
	if err != nil {
		return fmt.Errorf("Unrouted Exchange Declare: %w", err)
	}

	// Notification Exchange
	err = ch.ExchangeDeclare(
		NotificationExchange, // Name
		amqp.ExchangeDirect,  // Kind
		true,                 // Durable
		false,                // Auto delete
		false,                // Internal
		false,                // No wait
		amqp.Table{
			"alternate-exchange": UnroutedNotificationExchange,
		},
	)
	if err != nil {
		return fmt.Errorf("Exchange Declare: %w", err)
	}

	// Unroutable Notification Queue
	_, err = ch.QueueDeclare(
		UnroutedNotificationQueueName, // Queue Name
		true,                          // Durable
		false,                         // Auto delete
		false,                         // Exclusive
		false,                         // No Wait
		amqp.Table{
			"x-queue-type":           amqp.QueueTypeQuorum,
			"x-max-length":           UnroutedNotificationMaxLength,
			"x-message-ttl":          UnroutedNotificationMsgTTL.Milliseconds(),
			"x-dead-letter-exchange": NotificationDLExchange,
		},
	)
	if err != nil {
		return fmt.Errorf("Unrouted Queue Declare: %w", err)
	}

	// Binding Unrouted Exchange and Queue
	return ch.QueueBind(UnroutedNotificationQueueName, "*", UnroutedNotificationExchange, false, nil)
}

func (r *rabbitMQExt) PublishWithTTL(ctx context.Context, document string, message interface{}, ttl time.Duration) error {
	rabbitMqConfig := routingRabbitMqConfig(document)

	ch, err := r.getChannel()
	if err != nil {
		return fmt.Errorf("create channel: %v", err)
	}
	defer func() {
		if e := ch.Close(); e != nil {
			err = fmt.Errorf("close channel: %v", err)
		}
	}()

	// Declare the exchange
	err = ch.ExchangeDeclare(
		rabbitMqConfig.Exchange, // name
		amqp.ExchangeDirect,     // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		nil,                     // arguments
	)
	if err != nil {
		return fmt.Errorf("exchange declare (#1): %v", err)
	}

	// Declare the pending queue with a TTL and dead-letter settings
	pendingQueue, err := ch.QueueDeclare(
		rabbitMqConfig.DLQQueueName, // name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    rabbitMqConfig.Exchange,
			"x-dead-letter-routing-key": rabbitMqConfig.RoutingKey,
			"x-queue-type":              amqp.QueueTypeQuorum,
		},
	)
	if err != nil {
		return fmt.Errorf("queue declare (#1): %v", err)
	}

	// Bind the pending queue to the exchange with an empty routing key
	err = ch.QueueBind(
		pendingQueue.Name,            // queue name
		rabbitMqConfig.DLQRoutingKey, // routing key
		rabbitMqConfig.Exchange,      // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue binding (#1): %v", err)
	}

	// Declare the process queue
	processQueue, err := ch.QueueDeclare(
		rabbitMqConfig.QueueName, // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		amqp.Table{
			"x-queue-type": amqp.QueueTypeQuorum,
		}, // arguments
	)
	if err != nil {
		return fmt.Errorf("queue declare (#2): %v", err)
	}

	// Bind the process queue to the exchange with the routing key `settlement.processing`
	err = ch.QueueBind(
		processQueue.Name,         // queue name
		rabbitMqConfig.RoutingKey, // routing key
		rabbitMqConfig.Exchange,   // exchange
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue binding (#2): %v", err)
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when marshaling message", pdkLogger.Error(err))
		return err
	}

	publishing := amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent,
		Body:         jsonData,
		Expiration:   fmt.Sprintf("%d", ttl.Milliseconds()),
		Headers:      amqp.Table{},
		Timestamp:    time.Now().UTC(),
	}

	if traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string); traceID != "" {
		publishing.Headers[HeaderTraceId] = traceID
	}

	// Publish a message to the exchange (this message will go to the pending queue)
	err = ch.PublishWithContext(
		ctx,
		rabbitMqConfig.Exchange,      // exchange
		rabbitMqConfig.DLQRoutingKey, // routing key
		false,                        // mandatory
		false,                        // immediate
		publishing,
	)
	if err != nil {
		return fmt.Errorf("publish message (#1): %v", err)
	}

	return nil
}

// Example:
//
// buf := new(bytes.Buffer)
// buf.ReadFrom(r.Body)
//
// event, ch, err := rabbitmqExt.PublishAndWaitReply(ctx, "your-queue-name", buf.Bytes())
// if err != nil {
//   /* Handling errors when publishing messages */
// }
// defer ch.Close() // Don't forget to close the channel when not in use.
//
// to, cancel := context.WithTimeout(ctx, 60*time.Second)
// defer cancel()

// select {
// case <-to.Done():
//
//	/* Waiting time has been reached */
//
// case msg := <-event.Delivery:
//
//	if msg.Body == nil {
/*     // Handling when a channel or connection is closed
//  }
//	// Reply successfully received */
// }

func (r *rabbitMQExt) PublishWithDelay(ctx context.Context, document string, message interface{}, delay time.Duration) error {
	rabbitMqConfig := routingRabbitMqConfig(document)

	ch, err := r.getChannel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		rabbitMqConfig.Exchange+"_delayed",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-delayed-type": "direct", // underlying exchange type is direct
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when declaring delayed exchange", pdkLogger.Error(err))
		return err
	}

	args := amqp.Table{
		"x-queue-type": amqp.QueueTypeQuorum,
	}
	if rabbitMqConfig.DLQExchange != "" {
		args["x-dead-letter-exchange"] = rabbitMqConfig.DLQExchange
		args["x-dead-letter-routing-key"] = rabbitMqConfig.RoutingKey
		if rabbitMqConfig.DLQRoutingKey != "" {
			args["x-dead-letter-routing-key"] = rabbitMqConfig.DLQRoutingKey
		}
	}

	_, err = ch.QueueDeclare(
		rabbitMqConfig.QueueName,
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return fmt.Errorf("queue declare: %v", err)
	}

	err = ch.QueueBind(
		rabbitMqConfig.QueueName,
		rabbitMqConfig.RoutingKey,
		rabbitMqConfig.Exchange+"_delayed",
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("queue binding: %v", err)
	}

	var reqB []byte
	switch msg := message.(type) {
	case nil:
		return errors.New("message is nil")
	case []byte:
		reqB = msg
	default:
		// ensure message is marshalable
		reqB, err = json.Marshal(message)
		if err != nil {
			return fmt.Errorf("json marshal: %v", err)
		}
	}

	headers := amqp.Table{
		PluginHeaderXDelay: delay.Milliseconds(),
	}

	if retryCount, ok := ctx.Value(constant.CtxRabbitMQRetryCount).(int32); ok {
		headers[HeaderXRetryCount] = retryCount + 1
	}

	publishing := amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent,
		Body:         reqB,
		Headers:      headers,
		Timestamp:    time.Now().UTC(),
	}

	if traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string); traceID != "" {
		publishing.Headers[HeaderTraceId] = traceID
	}

	err = ch.PublishWithContext(
		ctx,
		rabbitMqConfig.Exchange+"_delayed",
		rabbitMqConfig.RoutingKey,
		false,
		false,
		publishing,
	)
	if err != nil {
		return fmt.Errorf("publish message: %v", err)
	}

	return nil
}

func (r *rabbitMQExt) PublishAndWaitReply(ctx context.Context, document string, message interface{}) (event *amqp.Event, close io.Closer, err error) {

	ch, err := r.getChannel()
	if err != nil {
		return nil, nil, fmt.Errorf("Get Channel: %w", err)
	}
	defer func() {
		if err != nil {
			_ = ch.Close()
		}
	}()

	cfg := routingRabbitMqConfig(document)

	if _, ok := r.once.directReplyTo.Load(document); !ok {

		_, err = ch.QueueDeclare(cfg.QueueName, true, false, false, false, amqp.Table{
			"x-queue-type": amqp.QueueTypeQuorum,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("Queue Declare: %w", err)
		}
		r.once.directReplyTo.Store(document, cfg)
	}

	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	contentType := "text/plain"
	switch data := message.(type) {
	case []byte:
		_, _ = buf.Write(data)

	default:
		_ = json.NewEncoder(buf).Encode(data)
		contentType = constant.MIMEApplicationJSON
	}

	_ = ch.Qos(1, 0, false)

	msgs, err := ch.Consume(ReplyToQueueName, "", true, true, false, false, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("Wait Reply: %w", err)
	}

	err = ch.PublishWithContext(
		ctx,
		"",            // Default Exchange
		cfg.QueueName, // Queue Name
		false,         // Mandatory
		false,         // Immediate
		amqp.Publishing{
			ContentType:   contentType,
			Body:          buf.Bytes(),
			ReplyTo:       ReplyToQueueName,
			CorrelationId: uuid.NewString(),
		},
	)
	if err != nil {
		return nil, nil, fmt.Errorf("Publish Message: %w", err)
	}
	return msgs, ch, nil
}

func (r *rabbitMQExt) PublishToReplyQueue(ctx context.Context, replyToAddress string, payload amqp.Publishing) error {
	ch, err := r.getChannel()
	if err != nil {
		return fmt.Errorf("Get Channel: %w", err)
	}
	defer ch.Close()

	return ch.PublishWithContext(ctx, "", replyToAddress, false, false, payload)
}
