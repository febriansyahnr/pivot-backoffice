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
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/virtualAccount"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
)

func (r *snapCoreRepository) CreateVirtualAccount(ctx context.Context, request snapCoreModel.CreateVirtualAccountRequest) (*snapCoreModel.CreateVirtualAccountResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/CreateVirtualAccount")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/virtual-account/create", r.config.SnapCoreConfig.BaseUrl)

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
			constant.HeaderXMerchantID:         request.MerchantID,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request create virtual account", logger.Error(err))
		return nil, err
	}

	var resp snapCoreModel.CreateVirtualAccountResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read create virtual account response body", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errMsg = resp.Error.Message
	}

	if errType, mapped := mapPartnerHTTPStatusToErrorType(statusCode); mapped {
		err = pkgErrors.New(errType, errors.New(errMsg))
		if statusCode >= http.StatusInternalServerError {
			r.logger.Error(ctx, fmt.Sprintf("got error %d when create virtual account, errorCode %s", statusCode, resp.Code), logger.Error(err))
		} else {
			r.logger.Warn(ctx, fmt.Sprintf("got error %d when create virtual account, errorCode %s", statusCode, resp.Code), logger.Any("errorResponse", err))
		}
		return nil, err
	}

	r.logger.Info(ctx, "response from create virtual account", logger.String("body", string(response)))
	return &resp.Data, nil
}
