package paymentConsumerController

import (
	"context"
	"encoding/json"
	"time"

	model "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"

	"github.com/paper-indonesia/pdk/v2/logger"
)

func (p *paymentConsumer) VCCTerminalSubmitCharge(ctx context.Context, body []byte, _ string) (err error) {
	ctx, segment := otelTracer.Start(ctx, "port/consumer/payment/VCCTerminalSubmitCharge")
	defer segment.End()

	var message any = string(body)

	start := time.Now()
	defer func() {
		duration := time.Since(start)

		p.logger.Info(
			ctx, "Charging VCC terminal transaction",
			logger.Any("charge", message), logger.Int64("duration", duration.Milliseconds()), logger.String("durationHuman", duration.String()), logger.Error(err),
		)
	}()

	request := model.VCCTerminalChargeMessage{}
	if err = json.Unmarshal(body, &request); err != nil {
		return pkgErrs.NewNonRetryableError(err)
	}
	message = request

	return p.paymentSvc.VCCTerminalSubmitCharge(ctx, request)
}
