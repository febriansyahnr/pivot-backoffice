package xbPayoutController

import (
	"mime/multipart"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/errors"
)

func (c *xbPayoutController) UploadUnderlyingDocument(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/UploadUnderlyingDocument")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantID := user.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrValidation, constant.ErrIdIsRequired))
		return
	}

	var (
		err     error
		f       multipart.File
		request xbModel.UploadUnderlyingDocumentRequest
	)

	if f, request.Document, err = r.FormFile("document"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, err))
		return
	}
	defer f.Close()

	if !constant.IsUnderlyingDocumentValidToUpload(request.Document.Filename) {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, errors.New("document format is not supported")))
		return
	}

	request.PayoutId = id
	request.MerchantId = merchantID

	resp, err := c.xbPayoutSvc.UploadUnderlyingDocument(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
