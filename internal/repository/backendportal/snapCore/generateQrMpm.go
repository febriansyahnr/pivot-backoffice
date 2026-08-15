package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/qr"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
)

func (r *snapCoreRepository) GenerateQrMpm(ctx context.Context, request snapCoreModel.GenerateQrMpmRequest) (*snapCoreModel.GenerateQrMpmResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/GenerateQrMpm")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/qr-mpm/generate", r.config.SnapCoreConfig.BaseUrl)

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
		r.logger.Error(ctx, "error when do request generate qr mpm", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.GenerateQrMpmResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read generate qr mpm response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if errType, mapped := mapPartnerHTTPStatusToErrorType(statusCode); mapped {
		err = pkgErrors.New(errType, errors.New(errMsg))
		if statusCode >= http.StatusInternalServerError {
			r.logger.Error(ctx, fmt.Sprintf("got error %d when generate qr mpm, errorCode %s", statusCode, resp.Code), logger.Error(err))
		} else {
			r.logger.Warn(ctx, fmt.Sprintf("got error %d when generate qr mpm, errorCode %s", statusCode, resp.Code), logger.Any("errorResponse", err))
		}
		return nil, err
	}

	r.logger.Info(ctx, "response from generate qr mpm", logger.String("body", string(response)))
	return &resp.Data, nil
}
