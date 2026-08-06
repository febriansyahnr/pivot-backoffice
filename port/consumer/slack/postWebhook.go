package slackConsumerController

import (
	"context"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/pkg/slackExt"
	"github.com/paper-indonesia/pdk/v2/logger"

	"google.golang.org/protobuf/proto"
)

// ProcessSlackPostWebhook implements consumer.ISlackConsumer.
func (s *SlackConsumer) ProcessSlackPostWebhook(ctx context.Context, body []byte, channel string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/slack/ProcessSlackPostWebhook")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error(ctx, "Panic recovery from ProcessSlackPostWebhook", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	payload := &pb.PostWebhookCmd{}
	if err := proto.Unmarshal(body, payload); err != nil {
		return constant.ErrUnmarshalProto
	}

	request := &slackExt.PostWebhookCmd{
		Color:  slackExt.SlackColor(payload.Color.String()),
		URL:    payload.URL,
		Title:  payload.Title,
		Fields: make([]slackExt.AttachmentField, len(payload.Fields)),
	}
	for i, field := range payload.Fields {
		request.Fields[i] = slackExt.AttachmentField{
			Title: field.Title,
			Value: field.Value,
			Short: field.Short,
		}
	}
	return s.slack.PostWebhook(ctx, request)
}
