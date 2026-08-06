package reconciliation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/reconciliation"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// ReconciliationProcess implements consumer.IReconciliationConsumer.
func (c *ReconciliationController) ReconciliationProcess(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/reconciliation/ReconciliationProcess")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from ReconciliationProcess", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	type Request struct {
		UUID string `json:"uuid"`
	}

	var req Request

	err := json.Unmarshal(body, &req)
	if err != nil {
		return err
	}

	if err := c.reconSvc.ProcessFile(ctx, req.UUID); err != nil {
		return err
	}

	return nil
}

func (c *ReconciliationController) SnapCoreTransferReconcile(ctx context.Context, body []byte, _ string) error {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/reconciliation/SnapCoreTransferReconcile")
	defer segment.End()

	defer func() {
		if r := recover(); r != nil {
			c.logger.Error(ctx, "Panic recovery from SnapCoreTransferReconcile", logger.Error(fmt.Errorf("%v", r)))
		}
	}()

	var req reconciliation.ReconciliationPayout

	err := json.Unmarshal(body, &req)
	if err != nil {
		return err
	}

	req.ProcessorReferenceName = constant.SnapCoreProcessor

	if err := c.reconSvc.ProcessPayoutRecon(ctx, &req); err != nil {
		return err
	}

	return nil
}
