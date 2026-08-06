package creditcardCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	creditcardCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/creditcardCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *creditcardCoreProcessorRepository) BlockCard(ctx context.Context, request *creditcardCoreProcessorModel.BlockCardRequest) error {
	ctx, segment := otelTracer.Start(ctx, "repository/creditcardCoreProcessor/BlockCard")
	defer segment.End()

	url := fmt.Sprintf("%s/crm/v1/card/block", r.config.CreditcardCoreProcessorConfig.BaseUrl)
	r.logger.Info(ctx, "Block Card", logger.String("url", url), logger.String("cardUuid", request.CardUUID))

	response, statusCode, err := r.httpRequest.PUT(ctx, url, request, nil)
	if err != nil {
		r.logger.Error(ctx, fmt.Sprintf("Failed to block card %s", request.CardUUID), logger.Error(err))
		return err
	}

	// Parse response to get error details if any
	var resp struct {
		Code  string      `json:"code"`
		Error interface{} `json:"error"`
	}

	var errMsg string
	if err := json.Unmarshal(response, &resp); err == nil && resp.Error != nil {
		errMsg = resp.Error.(string)
	} else {
		errB, _ := json.Marshal(response)
		errMsg = string(errB)
	}

	if statusCode >= 400 && statusCode < 500 {
		err = pkgErrors.New(httpResponse.HttpErrRequest, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 400 when block card, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	if statusCode >= 500 {
		err = pkgErrors.New(httpResponse.HttpErrInternal, errors.New(errMsg))
		r.logger.Error(ctx, fmt.Sprintf("got error 500 when block card, errorCode %s", resp.Code), logger.Error(err))
		return err
	}

	r.logger.Info(ctx, fmt.Sprintf("Successfully blocked card %s", request.CardUUID))
	return nil
}