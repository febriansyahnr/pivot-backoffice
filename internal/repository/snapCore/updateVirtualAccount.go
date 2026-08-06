package snapCoreRepository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) UpdateVirtualAccount(
	ctx context.Context,
	request snapCoreModel.UpdateVirtualAccountRequest) (*snapCoreModel.UpdateVirtualAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/UpdateVirtualAccount")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/update", r.config.SnapCoreConfig.BaseUrl)

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
		r.logger.Error(ctx, "error when do request update virtual account", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.UpdateVirtualAccountResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read update virtual account response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when update virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = fmt.Errorf("%s", errMsg)
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when update virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "response from update virtual account", logger.String("body", string(response)))
	return &resp.Data, nil
}
