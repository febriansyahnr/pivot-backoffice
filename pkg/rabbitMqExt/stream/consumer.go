package stream

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/newrelic/go-agent/v3/newrelic"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var consumerTracer = otel.Tracer("StreamConsumer")

func (c *Client) ReadMessage(ctx context.Context, cfg ReadMessageConfig) error {

	if cfg.ConsumerName == "" {
		return errors.New("consumer name is required")

	} else if cfg.StreamQueueName == "" {
		return errors.New("stream queue name is required")

	} else if cfg.Handler == nil {
		return errors.New("consumer handler func must be set")
	}

	cfg.setDefaults()

	offset, err := c.env.QueryOffset(cfg.ConsumerName, cfg.StreamQueueName)
	if err != nil && !strings.Contains(err.Error(), "Offset not found") {
		return fmt.Errorf("query offset: %w", err)
	}

	if offset > 0 {
		offset++
	}

	for {
		log.Printf("Attempting to connect to stream queue %s.\n", cfg.StreamQueueName)

		consumer, err := c.env.NewConsumer(
			cfg.StreamQueueName,
			c.wrapMessageHandler(cfg.RetryCount, cfg.RetryDelay, cfg.Handler),
			stream.NewConsumerOptions().
				SetConsumerName(cfg.ConsumerName).
				SetAutoCommit(stream.NewAutoCommitStrategy().
					SetCountBeforeStorage(cfg.CommitSize).
					SetFlushInterval(cfg.CommitInterval),
				).
				SetOffset(stream.OffsetSpecification{}.Offset(offset)),
		)
		if err != nil {
			return fmt.Errorf("failed to create consumer: %w", err)
		}
		log.Printf("Connection to stream queue %s established, ready to consume messages.\n", cfg.StreamQueueName)

		select {
		case <-ctx.Done():
			log.Printf("Terminate %s stream consumer.\n", cfg.StreamQueueName)

			if err := consumer.Close(); err != nil {
				log.Printf("Failed to close consumer connection, error=%s.\n", err)
			}
			return ctx.Err()

		case closeEvent := <-consumer.NotifyClose():
			// Ensure the consumer is completely closed.
			_ = consumer.Close()

			log.Printf("Consumer close notification received, name=%s reason=%s. Reconnecting in %s.\n", closeEvent.Name, closeEvent.Reason, cfg.ReconnectDelay.String())

			time.Sleep(cfg.ReconnectDelay)
			if ctx.Err() != nil {
				return ctx.Err()
			}
			continue
		}
	}
}

func (c *Client) Close() error {
	if c.env != nil {
		if err := c.env.Close(); err != nil {
			return err
		}
	}

	log.Println("Stream consumer stopped gracefully")

	return nil
}

func (c *Client) wrapMessageHandler(attempts int, sleep time.Duration, handler func(context.Context, []byte) error) stream.MessagesHandler {
	return func(consumerContext stream.ConsumerContext, message *amqp.Message) {
		ctx, span := consumerTracer.Start(context.Background(), "Consume Stream Name: "+consumerContext.Consumer.GetStreamName(), trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()

		txn := c.cfg.NR.GetApp().StartTransaction("Consume Stream Name: " + consumerContext.Consumer.GetStreamName())
		defer txn.End()

		txn.AddAttribute(newrelic.AttributeMessageSystem, "rabbitmq")
		txn.AddAttribute(newrelic.AttributeMessageQueueName, consumerContext.Consumer.GetStreamName())

		ctx = newrelic.NewContext(ctx, txn)
		ctx = context.WithValue(ctx, pdkConst.CtxTraceIdKey, uuid.NewString())

		body := message.GetData()
		if err := retryOnError(attempts, sleep, func() error { return handler(ctx, body) }); err != nil {
			log.Printf("All retries failed for stream %s, latest error=%s\n", consumerContext.Consumer.GetStreamName(), err)
		}
	}
}

func retryOnError(attempts int, sleep time.Duration, fn func() error) (err error) {
	attempts++
	for i := range attempts {
		if err = fn(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(sleep)
		}
	}
	return err
}
