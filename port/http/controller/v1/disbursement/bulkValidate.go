package disbursementController

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *Controller) BulkValidate(w http.ResponseWriter, r *http.Request) {

	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/BulkValidate")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}
	defer file.Close()

	request := &disbursementModel.BulkPreviewRequest{
		MerchantId: user.MerchantId,
		File:       file,
	}
	if previewResult, err := c.disbursementSvc.BulkValidate(ctx, request); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, previewResult)
	}
}
