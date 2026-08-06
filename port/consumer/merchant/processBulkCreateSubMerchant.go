package merchantConsumerController

import (
	"context"
	"encoding/json"
	"fmt"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *merchantConsumer) ProcessBulkCreateSubMerchant(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/merchant/ProcessBulkCreateSubMerchant")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from ProcessBulkCreateSubMerchant", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var req merchantModel.ProcessBulkCreateSubMerchantRequest
	if err := json.Unmarshal(body, &req); err != nil {
		c.logger.Error(ctx, "error when unmarshal body", logger.Error(err), logger.ByteString("body", body))
		return err
	}

	if err := c.merchantSvc.ProcessBulkCreateSubMerchant(ctx, &req); err != nil {
		c.logger.Error(ctx, "error when process bulk create submerchant", logger.Error(err))
		return err
	}
	return nil
}
