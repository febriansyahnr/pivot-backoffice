package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	snapCoreModel "github.com/paper-indonesia/pivot-backoffice/internal/model/snapCore/bankTransfer"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	pdkLog "github.com/paper-indonesia/pdk/v2/logger"
)

// CheckAllowedToRetry checks if a transaction can be retried
func (r *snapCoreRepository) CheckAllowedToRetry(ctx context.Context, request snapCoreModel.CheckAllowedToRetryRequest) (*snapCoreModel.CheckAllowedToRetryResponse, error) {
	ctx, span := otelTracer.Start(ctx, "internal/repository/snapCore/CheckAllowedToRetry")
	defer span.End()

	var (
		url      string
		response snapCoreModel.CheckAllowedToRetryResponse
	)

	url = r.config.SnapCoreConfig.BaseUrl + "/api/v1.0/internal/bank-transfer/check-allowed-to-retry"
	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	responseBytes, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			"X-MERCHANT-ID":                    request.MerchantId,
			constant.HeaderXInternalServiceKey: r.secret.SnapCoreSecret.InternalServiceKey,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request bank transfer", pdkLog.Error(err))
		return nil, err
	}
	if err = json.Unmarshal(responseBytes, &response); err != nil {
		r.logger.Error(
			ctx, errInvalidBankTransferRespFormat,
			pdkLog.Error(err), pdkLog.String("transactionId", request.ExternalID),
			pdkLog.String("merchantId", request.MerchantId), pdkLog.Any("request", request), pdkLog.ByteString("response", responseBytes),
		)

		// For further investigation process, this transaction will be made PENDING status.
		return &snapCoreModel.CheckAllowedToRetryResponse{
			Allowed: false,
			Reason:  errInvalidBankTransferRespFormat,
		}, errors.New(errInvalidBankTransferRespFormat)
	}
	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New("internal server error"))
		r.logger.Error(ctx, "got error 500 when do request bank transfer", pdkLog.Error(err))
		return &response, err

	} else if statusCode == http.StatusConflict {
		r.logger.Warn(
			ctx, "Double disbursement indication found, bank transfer response 409 Conflict", pdkLog.Any("request", request), pdkLog.Any("response", response),
		)
		return nil, constant.ErrDoubleDisbursementIndication

	} else if statusCode >= 400 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New("request failed"))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when do request bank transfer, errorCode %s", response.Reason), pdkLog.Error(err))
		return &response, err
	}

	return &response, nil
}
