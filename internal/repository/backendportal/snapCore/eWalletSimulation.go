package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/ewallet"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (r *snapCoreRepository) EWalletPaymentSimulation(ctx context.Context, request *ewallet.EWalletPaymentSimulationRequest) (*ewallet.EWalletPaymentSimulationResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/CreateEWalletPaymentLink")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/debit/payment-simulation", r.config.SnapCoreConfig.BaseUrl)

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request to ewallet payment simulation", logger.Error(err))
		return nil, err
	}

	var resp ewallet.SnapCoreEWalletPaymentSimulationResponse

	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read ewallet payment simulation response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "Response from create ewallet payment simulation", logger.Any("request", request), logger.Any("response", map[string]any{
		"status": statusCode, "body": string(response),
	}))

	if statusCode >= 500 {
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, constant.ErrPaymentPartnerInGeneral)

	} else if statusCode >= 400 {
		return nil, pkgErrors.New(httpResponse.HttpErrRequest, constant.ErrPaymentPartnerInGeneral)
	}

	return resp.Data, nil
}
