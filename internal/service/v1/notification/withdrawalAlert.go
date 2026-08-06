package notificationService

import (
	"context"
	"fmt"

	"github.com/gogo/protobuf/proto"
	slackPb "github.com/paper-indonesia/pivot-backoffice/internal/model/proto/messages/slack"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/withdrawal"
	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (s *notificationService) SendFailedWithdrawalAlert(ctx context.Context, request *withdrawal.FailedWithdrawalAlertRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/service/v1/notification/SendFailedWithdrawalAlert")
	defer segment.End()

	fields := []*slackPb.AttachmentField{
		{Title: "Withdrawal ID", Value: request.WithdrawalID, Short: true},
		{Title: "Merchant ID", Value: request.MerchantID, Short: true},
		{Title: "Balance Name", Value: request.BalanceName, Short: true},
		{Title: "Beneficiary Account Name", Value: request.BeneficiaryAccountName, Short: true},
		{Title: "Beneficiary Account No", Value: request.BeneficiaryAccountNo, Short: true},
		{Title: "Beneficiary Account Bank", Value: request.BeneficiaryAccountBankName, Short: true},
		{Title: "Amount", Value: fmt.Sprintf("Rp %s", util.ConvertFloatToCurrency(request.Amount)), Short: true},
		{Title: "Withdraw Type", Value: request.WithdrawType, Short: true},
		{Title: "Status", Value: request.Status, Short: true},
		{Title: "Failure Reason", Value: request.Reason, Short: true},
	}

	slackMsg := &slackPb.PostWebhookCmd{
		URL:    s.config.SlackConfig.WithdrawalAlertWebHookURL,
		Color:  slackPb.Color_GOOD,
		Title:  "<!subteam^S0738JK0LP9> " + request.AlertTitle,
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
