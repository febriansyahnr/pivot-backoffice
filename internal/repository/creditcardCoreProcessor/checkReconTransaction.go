package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

// CheckReconTransaction implements repository.ICreditcardCoreProcessorRepository.
func (r *creditcardCoreProcessorRepository) CheckReconTransaction(ctx context.Context, request *creditcardCoreProcessorModel.AutoReconTrxRequest) (*creditcardCoreProcessorModel.AutoReconTrxResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/CheckReconTransaction")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1.0/internal/auto-recon/transaction", r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "Check Recon Transaction", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request check recon transaction to credit card core processor", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.AutoReconResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read check recon transaction response body from credit card core processor", logger.Error(err))
		return nil, err
	}

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when check recon transaction from credit card core processor, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when check recon transaction from credit card core processor, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "response from check recon transaction credit card core processor", logger.String("body", string(response)))
	return &resp.Data, nil
}
