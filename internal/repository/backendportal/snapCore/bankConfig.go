package snapCoreRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pdk/v2/logger"
	snapCoreBankConfigModel "github.com/paper-indonesia/pivot-backoffice/internal/model/backendportal/snapCore/bankConfig"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (r *snapCoreRepository) GetBankCodeList(ctx context.Context, filter *snapCoreBankConfigModel.GetBankCodeListRequest) (*snapCoreBankConfigModel.BankCodeListResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/snapCore/FindBankTransferByExternalID")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/bank-config/bank-codes?transferType=%s&isActive=%d", r.config.SnapCoreConfig.BaseUrl, filter.TransferType, filter.IsActive)

	r.logger.Info(ctx, "GetBankCodeList", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			"X-Internal-Service-Key": r.secret.SnapCoreSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get bank code list", logger.Error(err))
		return nil, err
	}

	var resp snapCoreBankConfigModel.BankCodeListResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read get bank code list response body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "GetBankCodeList Response", logger.String("response", string(response)))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when get bank code list, errorCode %s", resp.Code), logger.Error(err))
		return &resp.Data, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when get bank code list, errorCode %s", resp.Code), logger.Error(err))
		return &resp.Data, err
	}

	return &resp.Data, nil
}
