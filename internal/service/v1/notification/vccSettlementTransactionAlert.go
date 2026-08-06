package notificationService

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogo/protobuf/proto"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *notificationService) SendVccSettlementTransactionAlert(ctx context.Context, request *vccSettlement.VccTransactionInquiryAlert) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/notification/SendVccSettlementTransactionAlert")
	defer segment.End()

	fields := []*slackPb.AttachmentField{
		{Title: "RcnID", Value: request.RcnId, Short: true},
		{Title: "Description", Value: request.Description, Short: true},
		{Title: "Posting Date", Value: request.PostingDate, Short: true},
	}
	recipientList := make([]string, len(request.Recipient))
	for i, val := range request.Recipient {
		recipientList[i] = fmt.Sprintf("<!subteam^%s>", val)
	}

	slackMsg := &slackPb.PostWebhookCmd{
		URL:    s.config.SlackConfig.VccSettlementTransactionInquiryAlertWebhookURL,
		Color:  slackPb.Color_GOOD,
		Title:  strings.Join(recipientList, ",") + " " + request.Title,
		Fields: fields,
	}
	rawSlackMessage, _ := proto.Marshal(slackMsg)
	err := s.rabbitMqExt.Publish(ctx, rabbitMqExt.SlackPostWebhookRoutingKey, nil, rawSlackMessage)
	if err != nil {
		s.logger.Error(ctx, "error when publish to slack queue", logger.Error(err))
		return err
	}

	return nil
}
