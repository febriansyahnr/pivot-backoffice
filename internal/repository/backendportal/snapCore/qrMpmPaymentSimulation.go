package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapQrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qris"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
)

func (r *snapCoreRepository) QrMpmPaymentSimulation(ctx context.Context, data *snapQrisModel.QrMpmPaymentSimulationRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/QrMpmPaymentSimulation")
	defer segment.End()

	snapURL := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/payment-simulation", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx, snapURL, data,
		map[string]string{
			constant.HeaderXRequestId:          requestId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		return err
	}

	r.logger.Info(
		ctx, "response from QR MPM payment simulation",
		logger.String("PartnerReferenceNo", data.PartnerReferenceNo),
		logger.String("url", snapURL), logger.Int("statusCode", statusCode), logger.ByteString("response", response),
	)

	respBody := map[string]interface{}{}
	if err := json.Unmarshal([]byte(response), &respBody); err != nil {
		return err

	} else if statusCode >= 400 {
		return &validation.Fields{"snap": respBody}
	}
	return nil
}
