package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// CheckReconTransaction implements repository.ISnapCoreRepository.
func (r *snapCoreRepository) CheckReconTransaction(ctx context.Context, request *snapCoreModel.AutoReconTrxRequest) (*snapCoreModel.AutoReconTrxResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/GenerateQrMpm")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/auto-recon/transaction", r.config.SnapCoreConfig.BaseUrl)

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
		r.logger.Error(ctx, "error when do request check recon transaction", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.AutoReconResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read check recon transaction response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when check recon transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when check recon transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "response from check recon transaction", logger.String("body", string(response)))
	return &resp.Data, nil
}
