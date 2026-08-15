package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (r *snapCoreRepository) CancelQrMpm(ctx context.Context, qrisID string) (*snapCoreModel.CancelQrMpmResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/snapCore/CancelQrMpm")
	defer span.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/cancel/%s", r.config.SnapCoreConfig.BaseUrl, qrisID)
	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.DELETE(
		ctx,
		url,
		nil,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "CancelQrMpm | error when do request", logger.Error(err))
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		r.logger.Error(ctx, "CancelQrMpm | got unauthorized response", logger.Int("statusCode", statusCode))
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	r.logger.Info(ctx, "CancelQrMpm | response from cancel qr mpm", logger.Int("statusCode", statusCode), logger.ByteString("body", response))

	resp := &snapCoreModel.CancelQrMpmResponse{}
	if err = json.Unmarshal(response, resp); err != nil {
		r.logger.Error(ctx, "CancelQrMpm | error when read response body", logger.Error(err))
		return nil, constant.ErrInvalidUnmarshalJSON
	}

	if resp.Error != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Error.Message))
	}

	if statusCode >= http.StatusBadRequest {
		r.logger.Error(ctx, fmt.Sprintf("CancelQrMpm | got error %d , errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return resp.Data, nil
}
