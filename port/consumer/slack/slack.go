package slackConsumerController

import (
	"github.com/paper-indonesia/pivot-backoffice/pkg/slackExt"
	"github.com/paper-indonesia/pivot-backoffice/port/consumer"
	"github.com/paper-indonesia/pdk/v2/logger"
	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("SlackConsumer")

type SlackConsumer struct {
	slack  slackExt.SlackNotifier
	logger logger.ILogger
}

func New(
	slack slackExt.SlackNotifier,
	logger logger.ILogger,
) consumer.ISlackConsumer {
	return &SlackConsumer{
		slack:  slack,
		logger: logger,
	}
}
