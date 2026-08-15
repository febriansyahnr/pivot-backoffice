package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/payment"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validation"
)

func (r *snapCoreRepository) PublishPayment(ctx context.Context, request snapPaymentModel.PublishRequest) error {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/snapCore/PublishPayment")
	defer segment.End()

	snapURL := fmt.Sprintf("%s/api/v1.0/internal/payment/publish", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx, snapURL, request,
		map[string]string{
			constant.HeaderXRequestId:          requestId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		return err
	}

	r.logger.Info(
		ctx, "response from publish payment",
		logger.String("InternalReference", request.InternalReference),
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
