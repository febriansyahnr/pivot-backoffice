package slackExt

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/slack-go/slack"
	"go.opentelemetry.io/otel"
)

type SlackColor string
type AttachmentField = slack.AttachmentField

const ColorGood SlackColor = "good"

var otelTracer = otel.Tracer("SlackExt")

type PostWebhookCmd struct {
	Color  SlackColor              `json:"color"`
	URL    string                  `json:"url"`
	Title  string                  `json:"title"`
	Fields []slack.AttachmentField `json:"fields"`
}

type SlackNotifier interface {
	PostWebhook(ctx context.Context, cmd *PostWebhookCmd) error
}

type SlackExt struct {
	AuthorName    string
	AuthorSubname string
}

// PostWebhook implements SlackNotifier.
func (s *SlackExt) PostWebhook(ctx context.Context, cmd *PostWebhookCmd) error {
	_, segment := otelTracer.Start(ctx, "pkg/slackExt/PostWebhook")
	defer segment.End()

	attachment := slack.Attachment{
		Color:         string(cmd.Color),
		AuthorName:    s.AuthorName,
		AuthorSubname: s.AuthorSubname,
		Text:          cmd.Title,
		Ts:            json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
		Fields:        cmd.Fields,
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
	}
	return slack.PostWebhook(cmd.URL, &msg)
}

func NewSlackExt(authorName, authorSubname string) SlackNotifier {
	return &SlackExt{
		AuthorName:    authorName,
		AuthorSubname: authorSubname,
	}
}
