package rabbitMqExt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	messagingQueueModel "github.com/paper-indonesia/pivot-backoffice/internal/model/messagingQueue"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/callback"
	pdkLogger "github.com/paper-indonesia/pdk/v2/logger"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pdk/v2/amqp"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"google.golang.org/protobuf/proto"
)

type ModifyMessage func(*Publishing)

func (r *rabbitMQExt) PublishForSettlementProcess(ctx context.Context, payload messagingQueueModel.PublishSettlementProcessPayload) error {

	if payload.Day < 1 {
		return errors.New("settlement day must be greater than zero")
	}
	group := "time_plus_" + fmt.Sprintf("%02s", fmt.Sprintf("%d", payload.Day))

	routingKey := "settlement." + group + ".quorum"
	messageTTLPerQueue := time.Duration(payload.Day * 24 * int(time.Hour))

	ch, err := r.getChannel()
	if err != nil {
		return fmt.Errorf("create channel: %w", err)
	}
	defer func() {
		if e := ch.Close(); e != nil {
			r.logger.Error(ctx, "Failed while close rabbitmq channel", pdkLogger.Error(e))
		}
	}()

	// Main Exchange
	err = ch.ExchangeDeclare(
		SchedulingSettlementExchange, // Name
		amqp.ExchangeDirect,          // Type
		true,                         // Durable
		false,                        // Auto-deleted
		false,                        // Internal
		false,                        // No-wait
		nil,                          // Arguments
	)
	if err != nil {
		return fmt.Errorf("exchange declare: %w", err)
	}

	// A queue used to temporarily hold messages to be processed
	pendingQueue, err := ch.QueueDeclare(
		fmt.Sprintf(SchedulingSettlementPendingQueueNameFmt, group), // Queue Name
		true,  // Durable
		false, // Delete when unused
		false, // Exclusive
		false, // No-wait
		amqp.Table{
			"x-dead-letter-exchange":    SchedulingSettlementExchange,
			"x-dead-letter-routing-key": SettlementProcessingRoutingKey,
			"x-queue-type":              amqp.QueueTypeQuorum,
			"x-message-ttl":             messageTTLPerQueue.Milliseconds(),
		},
	)
	if err != nil {
		return fmt.Errorf("queue declare for hold message: %w", err)
	}

	// A queue used to process messages when their settlement time arrives
	processQueue, err := ch.QueueDeclare(
		SchedulingSettlementProcessQueueName, // Name
		true,                                 // Durable
		false,                                // Delete when unused
		false,                                // Exclusive
		false,                                // No-wait
		amqp.Table{
			"x-queue-type": amqp.QueueTypeQuorum,
		}, // Arguments
	)
	if err != nil {
		return fmt.Errorf("queue declare for settlement process: %w", err)
	}

	binds := [2][2]string{
		// Queue name
		// Routing key
		{pendingQueue.Name, routingKey},
		{processQueue.Name, SettlementProcessingRoutingKey},
	}
	for i, bind := range binds {
		if err = ch.QueueBind(bind[0], bind[1], SchedulingSettlementExchange, false, nil); err != nil {
			return fmt.Errorf("queue binding (#%d): %w", i, err)
		}
	}

	message := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Headers:      amqp.Table{},
		Timestamp:    time.Now().UTC(),
	}

	switch content := payload.Payload.(type) {
	case []byte:
		message.Body = content
		message.ContentType = "text/plain"

	default:
		if message.Body, err = json.Marshal(content); err != nil {
			return fmt.Errorf("json marshal: %w", err)
		}
		message.ContentType = "application/json"
	}

	if traceID, _ := ctx.Value(pdkConst.CtxTraceIdKey).(string); traceID != "" {
		message.Headers[HeaderTraceId] = traceID
	}

	if payload.ModifyMessage != nil {
		payload.ModifyMessage(&message)
	}

	return ch.PublishWithContext(
		ctx,
		SchedulingSettlementExchange, // Exchange
		routingKey,                   // Routing key
		false,                        // Mandatory
		false,                        // Immediate
		message,                      // Message content
	)
}

func (r *rabbitMQExt) PublishMerchantCallback(ctx context.Context, payload *callback.ProcessCallbackRequest) error {

	payloadBytes, err := proto.Marshal(payload)
	if err != nil {
		r.logger.Warn(ctx, "Failed while compiling proto message in preparation for sending callback", pdkLogger.Error(err))
	}

	routingKey := CallbackRoutingKey
	if constant.IsMerchantCallbackWorkflowEnabled(config.Environment()) {
		routingKey = WorkflowCallbackRoutingKey

		payloadBytes, _ = json.Marshal(map[string]string{
			"payload": base64.StdEncoding.EncodeToString(payloadBytes),
		})
	}

	messageId := uuid.NewString()

	ctx = context.WithValue(ctx, constant.CtxMessageId, messageId)
	r.logger.Info(ctx, "Publish merchant callback", pdkLogger.String("messageId", messageId), pdkLogger.ByteString("payload", payloadBytes))

	return r.Publish(ctx, routingKey, nil, payloadBytes)
}
