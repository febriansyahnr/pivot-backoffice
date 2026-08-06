package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) EncryptCardData(ctx context.Context, request *creditcardCoreProcessorModel.EncryptCardRequest) (*creditcardCoreProcessorModel.EncryptedCardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/EncryptCardData")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/internal/credit-cards/encrypt-card", r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "Encrypt card", logger.String("url", url))

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
		r.logger.Error(ctx, "error when do request encrypt card", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.BaseEncryptedCardResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when unmarshal encrypt card response", logger.Error(err))
		return nil, err
	}
	r.logger.Info(ctx, "Encrypt Card", logger.ByteString("response", response))

	if resp.Message != "" && statusCode != http.StatusOK {
		r.logger.Error(ctx, "error when encrypt card", logger.Any("response", resp), logger.String("error_code", resp.Code))
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Message))
		return nil, err
	}
	if statusCode >= 400 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(resp.Message))
		return nil, err
	}

	return &resp.Data, nil

}

func (r *creditcardCoreProcessorRepository) GetEncryptedCardData(ctx context.Context, merchantId, cardId string) (*creditcardCoreProcessorModel.EncryptedCardResponse, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/creditcardCoreProcessor/DecryptCardData")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/internal/credit-cards/encrypt-card/%s", r.config.CreditcardCoreProcessorConfig.BaseUrl, cardId)
	r.logger.Info(ctx, "Get encrypted card", logger.String("url", url))

	response, statusCode, err := r.httpRequest.GET(
		ctx,
		url,
		map[string]string{
			"X-MERCHANT-ID":                    merchantId,
			constant.HeaderXInternalServiceKey: r.config.CreditcardCoreProcessorConfig.BaseUrl,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request get encrypted card", logger.Error(err))
		return nil, err
	}

	var resp creditcardCoreProcessorModel.BaseEncryptedCardResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when unmarshal get encrypted card response", logger.Error(err))
		return nil, err
	}
	r.logger.Info(ctx, "Encrypt Card", logger.ByteString("response", response))

	if resp.Message != "" && statusCode != http.StatusOK {
		r.logger.Error(ctx, "error when get encrypted card", logger.Any("response", resp), logger.String("error_code", resp.Code))
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(resp.Message))
		return nil, err
	}
	if statusCode >= 400 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(resp.Message))
		return nil, err
	}

	return &resp.Data, nil

}
