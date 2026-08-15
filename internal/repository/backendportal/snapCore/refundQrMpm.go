package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapQrisModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/qris"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *snapCoreRepository) RefundQRMPM(ctx context.Context, req *snapQrisModel.QRMPMRefundRequest) (*snapQrisModel.RefundResponseData, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/snapCore/RefundQrMpm")
	defer span.End()
	r.logger.Info(ctx, "start refund qr mpm", logger.Any("payload", req))
	url := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/refund", r.config.SnapCoreConfig.BaseUrl)
	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		req,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do refund request", logger.Error(err))
		return nil, err
	}

	if statusCode == http.StatusUnauthorized {
		r.logger.Error(ctx, "got unauthorized response", logger.Int("statusCode", statusCode))
		return nil, pkgErrors.New(httpResponse.HttpErrInternal, constant.ErrInternalServerForUser)
	}

	r.logger.Info(ctx, "response from refund qr mpm", logger.Int("statusCode", statusCode), logger.ByteString("body", response))

	resp := &snapQrisModel.RefundQRMPMResponse{}
	if err = json.Unmarshal(response, resp); err != nil {
		r.logger.Error(ctx, "error when read response body", logger.Error(err))
		return nil, constant.ErrInvalidUnmarshalJSON
	}

	if resp.Error != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Error.Message))
	}

	if statusCode >= http.StatusBadRequest {
		r.logger.Error(ctx, fmt.Sprintf("got error %d , errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return resp.Data, nil
}
