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

func (r *snapCoreRepository) BlockVirtualAccount(ctx context.Context, request *snapCoreModel.BlockVirtualAccountRequest) ([]*snapCoreModel.BlockVirtualAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/BlockVirtualAccount")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/block", r.config.SnapCoreConfig.BaseUrl)

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
		r.logger.Error(ctx, "error when do request block virtual account", logger.Error(err))
		return nil, err
	}

	resp := &snapCoreModel.BlockVirtualAccountResponse{}
	if err = json.Unmarshal(response, resp); err != nil {
		r.logger.Error(ctx, "error when read block virtual account response body", logger.Error(err))
		return nil, constant.ErrInvalidUnmarshalJSON
	}

	if resp.Error != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Error.Message))
	}

	if statusCode >= 500 {
		r.logger.Error(ctx, fmt.Sprintf("got error 5xx when block virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err

	} else if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("got error 4xx when block virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return resp.Data, nil
}

func (r *snapCoreRepository) UnblockVirtualAccount(ctx context.Context, request *snapCoreModel.UnblockVirtualAccountRequest) ([]*snapCoreModel.UnblockVirtualAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/UnblockVirtualAccount")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/unblock", r.config.SnapCoreConfig.BaseUrl)

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
		r.logger.Error(ctx, "error when do request unblock virtual account", logger.Error(err))
		return nil, err
	}

	resp := &snapCoreModel.UnblockVirtualAccountResponse{}
	if err = json.Unmarshal(response, resp); err != nil {
		r.logger.Error(ctx, "error when read unblock virtual account response body", logger.Error(err))
		return nil, constant.ErrInvalidUnmarshalJSON
	}
	if resp.Error != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Error.Message))
	}

	if statusCode >= 500 {
		r.logger.Error(ctx, fmt.Sprintf("got error 5xx when unblock virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err

	} else if statusCode >= 400 {
		r.logger.Error(ctx, fmt.Sprintf("got error 4xx when unblock virtual account, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return resp.Data, nil
}
