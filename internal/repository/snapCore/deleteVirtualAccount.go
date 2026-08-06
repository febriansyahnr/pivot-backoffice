package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) DeleteVirtualAccount(ctx context.Context, request *snapCoreModel.DeleteVirtualAccountRequest) (*snapCoreModel.DeleteVirtualAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/DeleteVirtualAccount")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/delete", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.DELETE(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request delete virtual account", logger.Error(err))
		return nil, err
	}
	r.logger.Info(ctx, "response from delete virtual account", logger.Int("statusCode", statusCode), logger.ByteString("body", response))

	resp := &snapCoreModel.DeleteVirtualAccountResponse{}
	if err = json.Unmarshal(response, resp); err != nil {
		r.logger.Error(ctx, "error when read delete virtual account response body", logger.Error(err))
		return nil, constant.ErrInvalidUnmarshalJSON
	}

	if resp.Error != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Error.Message))
	}

	if statusCode >= 500 {
		r.logger.Error(ctx, fmt.Sprintf("got error 5xx when delete virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err

	} else if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("got error 4xx when delete virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}
	return resp.Data, nil
}
