package xbCoreProcessorRepository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"mime/multipart"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	xbCoreProcessorModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xbCoreProcessor"
	pdkConst "github.com/paper-indonesia/pdk/v2/constant"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (r *xbCoreProcessorRepository) UploadUnderlyingDocument(ctx context.Context, request *xbCoreProcessorModel.UploadUnderlyingDocumentRequest) (*xbCoreProcessorModel.UploadUnderlyingDocumentResponseData, error) {
	ctx, segment := otelTracer.Start(ctx, "internal/repository/v1/xbCoreProcessor/UploadUnderlyingDocument")
	defer segment.End()

	url := fmt.Sprintf("%s/api/v1/payout/%s/upload", r.config.XbCoreProcessorConfig.BaseUrl, request.XbPayoutId)
	r.logger.Info(ctx, "UploadUnderlyingDocument", logger.String("url", url), logger.Any("request", request))

	formData := map[string]string{}
	files := map[string]*multipart.FileHeader{}

	if request.Document != nil {
		files = map[string]*multipart.FileHeader{
			"document": request.Document,
		}
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
		r.logger.Error(ctx, "error when do request xb upload underlying document", logger.Error(err))
		return nil, err
	}

	var resp xbCoreProcessorModel.UploadUnderlyingDocumentResponse
	err = json.Unmarshal(response, &resp)
	if err != nil {
		r.logger.Error(ctx, "error when read xb upload underlying document body", logger.Error(err))
		return nil, err
	}

	r.logger.Info(ctx, "UploadUnderlyingDocument", logger.ByteString("response", response))

	if statusCode >= http.StatusBadRequest {
		err = mapXbStatusToError(statusCode, string(response))
		r.logger.Error(ctx, fmt.Sprintf("got error %d when xb upload underlying document, errorCode %s", statusCode, resp.Code), logger.Error(err))
		return nil, err
	}

	return &resp.Data, nil
}
