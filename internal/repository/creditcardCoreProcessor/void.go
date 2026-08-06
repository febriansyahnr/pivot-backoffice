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

func (r *creditcardCoreProcessorRepository) Void(
	ctx context.Context,
	request *creditcardCoreProcessorModel.VoidRequest,
) (*creditcardCoreProcessorModel.VoidResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/Void")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/credit-card/void", r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "Void", logger.String("url", url))

	response, statusCode, err := r.httpRequest.POST(
		ctx,
		url,
		request,
		map[string]string{
			"X-MERCHANT-ID":                    request.MerchantID,
			constant.HeaderXInternalServiceKey: r.secret.CreditcardCoreProcessorSecret.InternalServiceKey,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request void creditcard transaction", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.VoidResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read void creditcard transaction body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "Void", logger.ByteString("response", response))

	errMsg := resp.Message
	if resp.Error != nil {
		errB, _ := json.Marshal(resp.Error)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when void creditcard transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when void creditcard transaction, errorCode %s", resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
