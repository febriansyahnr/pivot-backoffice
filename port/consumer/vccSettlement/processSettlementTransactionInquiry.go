package vccsettlement

import (
	"context"
	"encoding/json"

	"github.com/paper-indonesia/pivot-backoffice/internal/model/vccSettlement"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *handler) ProcessSettlementTransactionInquiry(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/vccSettlement/ProcessSettlementTransactionInquiry")
	defer segment.End()

	var request vccSettlement.VccTransactionInquiryRequest
	err := json.Unmarshal(body, &request)
	if err != nil {
		c.logger.Error(ctx, "error unmarshal body request", logger.Error(err), logger.String("bodyPayload", string(body)))
		return err
	}

	err = c.service.ProcessRcnTransactionInquiry(ctx, &request)
	if err != nil {
		c.logger.Error(ctx, "error when process rcn transaction inquiry", logger.Error(err), logger.String("bodyPayload", string(body)))
		return err
	}

	c.logger.Info(ctx, "finish process rcn transaction inquiry", logger.String("bodyPayload", string(body)))
	return nil
}
