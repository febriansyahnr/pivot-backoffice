package v2InternalUnifiedPaymentController

import (
	"context"
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	unifiedPaymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/unifiedPayment"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func (c *paymentController) UploadProofOfPayment(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v2/internalController/unifiedPayment/UploadProofOfPayment")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}
	ctx = context.WithValue(ctx, constant.CtxExposeUnmappingRequestError, true)

	merchantID := merchantAuth.MerchantId
	if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
		merchantID = subMerchantId
	}

	paymentID := chi.URLParam(r, "uuid")
	if err := uuid.Validate(paymentID); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	maxFileSize := c.config.Investigation.MaxFileSizeMB * 1024 * 1024
	if err := r.ParseMultipartForm(1 << 20); err != nil { // 1MB in memory, rest goes to disk
		c.logger.Warn(ctx, "Failed to parse multipart form", logger.Error(err))
		if strings.Contains(err.Error(), constant.ContentTypeMultipartFormData) {
			response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
			return
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrFileTooLarge))
		return
	}

	file, fileHeader, err := r.FormFile("proofOfTransaction")
	if err != nil {
		c.logger.Warn(ctx, "Failed to get proof of transaction file", logger.Error(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}
	defer file.Close()

	reason := r.FormValue("reason")
	if reason == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidRequestPayload))
		return
	}

	fileExt := strings.ToLower(strings.Replace(filepath.Ext(fileHeader.Filename), ".", "", 1))
	if !slices.Contains(c.config.Investigation.AllowedFileExts, fileExt) {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidFileExtension))
		return
	}

	if fileHeader.Size > maxFileSize {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrFileTooLarge))
		return
	}

	request := &unifiedPaymentModel.UploadProofOfPaymentRequest{
		PaymentID:        paymentID,
		MerchantID:       merchantID,
		ProofOfPayment:   fileHeader,
		Reason:           reason,
		FileExtension:    fileExt,
		OriginalFileName: fileHeader.Filename,
	}

	resp, err := c.unifiedPaymentSvc.UploadProofOfPayment(ctx, request)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	response.SendOpenApiResponseOK(w, resp)
}
