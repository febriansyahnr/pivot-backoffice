package xbCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"mime/multipart"

	"github.com/paper-indonesia/pivot-backoffice/constant"

	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *xbCoreProcessorRepository) SubmitRfiDetails(ctx context.Context, request *xbCoreProcessorModel.SubmitRfiDetailsRequest) (*xbCoreProcessorModel.SubmitRfiDetailsResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/SubmitRfiDetails")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout/%s/rfi", r.config.XbCoreProcessorConfig.BaseUrl, request.PayoutId)
	r.logger.Info(ctx, "SubmitRfiDetails", logger.String("url", url), logger.Any("request", request))

	formData := map[string]string{
		"document_id": request.DocumentId,
		"comment":     request.Comment,
		"value":       request.Value,
	}
	files := map[string]*multipart.FileHeader{}

	if request.Document != nil {
		files = map[string]*multipart.FileHeader{
			"document": request.Document,
		}
		delete(formData, "value")
	}

	requestId, _ := ctx.Value(pdkConst.CtxRequestIdKey).(string)
	response, statusCode, err := r.httpRequest.POSTWithFormData(
		ctx,
		url,
		formData,
		files,
		map[string]string{
			constant.HeaderXInternalServiceKey: r.secret.XbCoreProcessorSecret.InternalServiceKey,
			constant.HeaderXMerchantId:         request.MerchantId,
			constant.XRequestIdKey:             requestId,
		},
	)
	if err != nil {
		r.logger.Error(ctx, "error when do request xb submit rfi details", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "SubmitRfiDetails", logger.ByteString("response", response))

	var resp xbCoreProcessorModel.SubmitRfiDetailsResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		err = pkgErrors.New(httpResponse.HttpErrInternal, err)
		r.logger.Error(ctx, "error when read xb submit rfi details body", logger.Error(err))
		return nil, err
	}

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb submit rfi details, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
