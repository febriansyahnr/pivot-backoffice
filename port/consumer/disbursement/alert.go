package disbursementConsumerController

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *DisbursementConsumer) PayoutTransactionAlertProcess(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/disbursement/PayoutTransactionAlertProcess")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from PayoutTransactionAlertProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	payload := &disbursementModel.PayoutTransactionAlertRequest{}
	if err = json.Unmarshal(body, payload); err != nil {
		return constant.ErrInvalidUnmarshalJSON
	}
	c.logger.Info(ctx, "Incoming message on PayoutTransactionAlertRequest", logger.Any("payload", payload))

	if err = c.disbursementSvc.ProcessPayoutAlert(ctx, payload); err != nil {
		return err
	}

	return nil
}
