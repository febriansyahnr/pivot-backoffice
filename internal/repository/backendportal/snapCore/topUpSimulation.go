package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/topUpSimulation"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (r *snapCoreRepository) TopUpSimulation(ctx context.Context, request snapCoreModel.TopupSimulationRequest) (*snapCoreModel.TopupSimulationResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/TopUpSimulation")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/topup-simulation", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request create virtual account", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.TopupSimulationResponse

	if err = json.Unmarshal(response, &resp); err != nil {
		r.logger.Error(ctx, "error when read create simulation top up VA response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	switch {
	case statusCode >= 500:
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when create simulation top up VA, errorCode %s", resp.Code), logger.Error(err))
	case statusCode >= 400:
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when create simulation top up VA, errorCode %s", resp.Code), logger.Error(err))
	default:
		r.logger.Info(ctx, "response from create simulation top up VA", logger.String("body", string(response)))
		return &resp.Data, nil
	}

	return nil, err
}
