package rabbitMqExt

import (
	"context"
	"fmt"
	"time"

	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/amqp"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var otelTracer = otel.Tracer("Consumer")

func (r *rabbitMQExt) Consume(signal context.Context, document string, exchangeType *string, process func(context.Context, []byte, string) error) error {

	rabbitMqConfig := routingRabbitMqConfig(document)

	rmqExchangeType := "topic"
	if exchangeType != nil {
		rmqExchangeType = *exchangeType
	}

	var exchangeArgs amqp.Table
	if rmqExchangeType == DelayedExchangeType {
		exchangeArgs = amqp.Table{
			"x-delayed-type": "direct",
		}
	}

	ch, err := r.getChannel()
	if err != nil {
		return err
	}

	args := amqp.Table{
		"x-queue-type": amqp.QueueTypeQuorum, // NOSONAR
	}
	if rabbitMqConfig.DLQExchange != "" {
		args["x-dead-letter-exchange"] = rabbitMqConfig.DLQExchange   // NOSONAR
		args["x-dead-letter-routing-key"] = rabbitMqConfig.RoutingKey // NOSONAR
	}

	err = ch.ExchangeDeclare(
		rabbitMqConfig.Exchange, // name
		rmqExchangeType,         // type
		true,                    // durable
		false,                   // auto-deleted
		false,                   // internal
		false,                   // no-wait
		exchangeArgs,            // arguments
	)
	if err != nil {
		r.logger.Error(signal, "error rabbitmq when exchange declare", pdkLogger.Error(err))
		return err
	}

	q, err := ch.QueueDeclare(
		rabbitMqConfig.QueueName, // Queue name
		true,                     // Durable (the queue will not survive server restarts)
		false,                    // Delete when unused
		false,                    // Exclusive (queue only accessible by the connection that declares it)
		false,                    // No-wait
		args,                     // Arguments
	)
	if err != nil {
		r.logger.Error(signal, "error rabbitmq when queue declare", pdkLogger.Error(err))
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
		r.logger.Error(signal, "error rabbitmq when queue bind", pdkLogger.Error(err))
		return err
	}

	// Declare dead letter, if dead-letter exists
	if rabbitMqConfig.DLQExchange != "" {
		err = ch.ExchangeDeclare(
			rabbitMqConfig.DLQExchange, // name
			rmqExchangeType,            // type
			true,                       // durable
			false,                      // auto-deleted
			false,                      // internal
			false,                      // no-wait
			exchangeArgs,               // arguments
		)
		if err != nil {
			r.logger.Error(signal, "error rabbitmq when exchange declare", pdkLogger.Error(err))
			return err
		}

		_, err = ch.QueueDeclare(
			rabbitMqConfig.DLQQueueName, // Queue name
			true,                        // Durable (the queue will not survive server restarts)
			false,                       // Delete when unused
			false,                       // Exclusive (queue only accessible by the connection that declares it)
			false,                       // No-wait
			amqp.Table{
				"x-queue-type": amqp.QueueTypeQuorum, // NOSONAR
			}, // Arguments
		)
		if err != nil {
			r.logger.Error(signal, "error rabbitmq when queue declare", pdkLogger.Error(err))
			return err
		}

		err = ch.QueueBind(
			rabbitMqConfig.DLQQueueName, // Queue name
			rabbitMqConfig.RoutingKey,   // Routing key
			rabbitMqConfig.DLQExchange,  // Exchange
			false,
			nil,
		)
		if err != nil {
			r.logger.Error(signal, "error rabbitmq when queue bind", pdkLogger.Error(err))
			return err
		}
	}

	messages, err := ch.ConsumeWithContext(
		signal, // main context for cancelation
		q.Name, // queue
		"",     // consumer tag
		false,  // auto-acknowledge
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		r.logger.Error(signal, "error rabbitmq when consume", pdkLogger.Error(err))
		return err
	}

	// TODO: where we should put this config
	needRetryMechanism := false

	defer ch.Close()
	defer ch.CloseConsume(messages)
	defer func() { r.logger.Info(signal, "Closing: "+document) }()

	for {
		select {
		case <-signal.Done():
			return nil

		case msg := <-messages.Delivery:
			if msg.Body == nil {
				continue
			}

			r.processMessage(rabbitMqConfig, needRetryMechanism, msg, process)
		}
	}
}

func (r *rabbitMQExt) processMessage(rabbitMqConfig RabbitMqExchangeConfig, needRetryMechanism bool, msg amqp.Delivery, process func(context.Context, []byte, string) error) {
	ctx, span := otelTracer.Start(context.Background(), "Consumer "+rabbitMqConfig.QueueName, trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	txn := r.nrApp.GetApp().StartTransaction("Consumer " + rabbitMqConfig.QueueName)
	defer txn.End()

	txn.AddAttribute(newrelic.AttributeMessageSystem, "rabbitmq")
	txn.AddAttribute(newrelic.AttributeMessageCorrelationID, msg.CorrelationId)
	txn.AddAttribute(newrelic.AttributeMessageHeaders, msg.Headers)
	txn.AddAttribute(newrelic.AttributeMessageExchangeType, rabbitMqConfig.Exchange)
	txn.AddAttribute(newrelic.AttributeMessageQueueName, rabbitMqConfig.QueueName)
	txn.AddAttribute(newrelic.AttributeMessageRoutingKey, rabbitMqConfig.RoutingKey)

	ctx = newrelic.NewContext(ctx, txn)

	attrs := []attribute.KeyValue{
		attribute.String("exchange", rabbitMqConfig.Exchange),
		attribute.String("queue", rabbitMqConfig.QueueName),
		attribute.String("routing", rabbitMqConfig.RoutingKey),
	}

	// Get traceID
	traceID, ok := msg.Headers[HeaderTraceId].(string)
	if !ok {
		traceID = uuid.NewString()
	}
	// Set the transaction in the context
	ctx = context.WithValue(ctx, constant.CtxRabbitMQStartTime, time.Now().UTC())
	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, traceID)

	if parentMerchantId, _ := msg.Headers[HeaderXParentMerchantId].(string); parentMerchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, parentMerchantId)
	}
	if derivedMerchantId, _ := msg.Headers[HeaderXDerivedMerchantId].(string); derivedMerchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxDerivedMerchantID, derivedMerchantId)
	}
	if merchantId, _ := msg.Headers[HeaderXMerchantId].(string); merchantId != "" {
		ctx = context.WithValue(ctx, constant.CtxMerchantIDKey, merchantId)
	}

	if retryCount, ok := msg.Headers[HeaderXRetryCount].(int32); ok {
		ctx = context.WithValue(ctx, constant.CtxRabbitMQRetryCount, retryCount)
	} else {
		ctx = context.WithValue(ctx, constant.CtxRabbitMQRetryCount, int32(0))
	}

	r.logger.Info(ctx, "consume message",
		pdkLogger.String("queueName", rabbitMqConfig.QueueName), pdkLogger.String("publishedAt", msg.Timestamp.Format(time.RFC3339)),
	)

	// Get ReplyTo when exists
	if msg.ReplyTo != "" {
		ctx = context.WithValue(ctx, constant.CtxRabbitMQReplyTo, msg.ReplyTo)
	}

	err := process(ctx, msg.Body, rabbitMqConfig.Channel)
	if err != nil {
		r.logger.Error(ctx, "error rabbitmq when processing message", pdkLogger.Error(err))

		span.RecordError(err)
		span.SetStatus(codes.Error, "An error occurred")

		if needRetryMechanism && msg.DeliveryTag < uint64(constant.MaxRetryMechanism) {
			// Use the utility function to safely convert uint64 to int64
			deliveryTagAttr := util.MustUint64ToInt64(msg.DeliveryTag)
			if tagInt64, ok := deliveryTagAttr.(int64); ok {
				attrs = append(attrs, attribute.Int64("delivery_tag", tagInt64))
			} else if tagStr, ok := deliveryTagAttr.(string); ok {
				attrs = append(attrs, attribute.String("delivery_tag", tagStr))
			}

			attrs = append(attrs, attribute.Bool("ack", false))
			span.SetAttributes(attrs...)

			if err := msg.Nack(false, true); err != nil {
				// Log error but continue processing
			}
			return
		}
	}

	attrs = append(attrs, attribute.Bool("ack", true))
	span.SetAttributes(attrs...)

	if err := msg.Ack(false); err != nil {
		// Log error but continue processing
	}
}

func (r *rabbitMQExt) ConsumeWithOpts(ctx context.Context, config *ConsumeOptsConfig, handler HandlerFunc, opts ...ChannelOpt) error {
	ch, err := r.getChannel()
	if err != nil {
		return fmt.Errorf("Get Channel: %w", err)
	}
	defer ch.Close()

	for _, opt := range opts {
		if err := opt(ch); err != nil {
			return fmt.Errorf("Set Options: %w", err)
		}
	}

	if config.RetryMechanism && config.RetryLimit == 0 {
		config.RetryLimit = 3
	}

	messages, err := ch.ConsumeWithContext(
		ctx, config.QueueName, "", false, false, false, false, config.Args,
	)
	if err != nil {
		return fmt.Errorf("Init Consumer: %w", err)
	}
	defer ch.CloseConsume(messages)
	defer func() { r.logger.Info(ctx, "Closing Consumer: "+config.QueueName) }()

	for {
		select {
		case <-ctx.Done():
			return nil

		case msg := <-messages.Delivery:
			if msg.Body == nil {
				continue
			}
			r.handlerProcess(config, handler, &msg)
		}
	}
}

func (r *rabbitMQExt) handlerProcess(config *ConsumeOptsConfig, handler HandlerFunc, msg *amqp.Delivery) {
	ctx, span := otelTracer.Start(context.Background(), "Processing Queue: "+config.QueueName, trace.WithSpanKind(trace.SpanKindConsumer))
	defer span.End()

	txn := r.nrApp.GetApp().StartTransaction("Consumer " + config.QueueName)
	defer txn.End()

	attrs := []attribute.KeyValue{
		attribute.String("queue", config.QueueName), attribute.String("routing_key", msg.RoutingKey),
	}

	txn.AddAttribute(newrelic.AttributeMessageSystem, "rabbitmq")
	txn.AddAttribute(newrelic.AttributeMessageCorrelationID, msg.CorrelationId)
	txn.AddAttribute(newrelic.AttributeMessageHeaders, msg.Headers)
	txn.AddAttribute(newrelic.AttributeMessageExchangeType, msg.Exchange)
	txn.AddAttribute(newrelic.AttributeMessageQueueName, config.QueueName)
	txn.AddAttribute(newrelic.AttributeMessageRoutingKey, msg.RoutingKey)

	ctx = newrelic.NewContext(ctx, txn)

	traceID, ok := msg.Headers[HeaderTraceId].(string)
	if !ok {
		traceID = uuid.NewString()
	}
	ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, traceID)

	if err := handler(ctx, msg); err != nil {

		span.RecordError(err)
		span.SetStatus(codes.Error, "An error occurred")

		if config.RetryMechanism && msg.DeliveryTag <= config.RetryLimit {

			if err := msg.Nack(false, true); err != nil {
				// Log error but continue processing
			}

			// Use the utility function to safely convert uint64 to int64
			var localAttrs []attribute.KeyValue
			deliveryTagAttr := util.MustUint64ToInt64(msg.DeliveryTag)
			if tagInt64, ok := deliveryTagAttr.(int64); ok {
				localAttrs = append(localAttrs, attribute.Int64("delivery_tag", tagInt64))
			} else if tagStr, ok := deliveryTagAttr.(string); ok {
				localAttrs = append(localAttrs, attribute.String("delivery_tag", tagStr))
			}

			localAttrs = append(localAttrs, attribute.Bool("ack", false))
			span.SetAttributes(localAttrs...)
			return
		}
	}
	if err := msg.Ack(false); err != nil {
		// Log error but continue processing
	}
	span.SetAttributes(append(attrs, attribute.Bool("ack", true))...)
}
