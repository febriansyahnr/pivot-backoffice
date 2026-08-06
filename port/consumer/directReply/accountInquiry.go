package directreply

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/paper-indonesia/pdk/v2/logger"

	routingProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/routingProcessor/accountInquiry"
)

func (c *directReply) AddressingWaitReplyAccountInquiry(ctx context.Context, body []byte, channel string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/directReply/AddressingWaitReplyAccountInquiry")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from AddressingWaitReplyAccountInquiry", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var payload routingProcessorModel.InquiryAccountResponseData

	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	return c.routingProcessorSvc.AddressingReplyToAccountInquiry(ctx, &payload)
}
