package subMerchant

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

func (c *SubMerchantController) DetailSubMerchantByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/api/v1/subMerchant/DetailSubMerchantByID")
	defer segment.End()

	merchant, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrRequiredSubmerchantId))
		return
	}

	subMerchant, err := c.merchantSvc.FindMerchantByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrInternal, constant.ErrFailedValidateSubMerchantParent))
		return

	} else if subMerchant == nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrSubMerchantNotFound))
		return

	} else if subMerchant.ParentID.String != merchant.MerchantId {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrIncorrectSubMerchantParent))
		return
	}
	response.SendApiResponseOK(w, subMerchant.ToResponse())
}
