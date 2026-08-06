package subMerchant

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetDailyTransactionLimit	godoc
// @Summary					Endpoint to view daily transaction limit of the sub merchant.
// @Description				It will get the daily transaction limit of the sub merchant based ont the merchant type.
// @ID						api-submerchant-get-daily-transaction-limit
// @Tags					API - Sub Merchant
// @Accept					json
// @Produce					json
// @Param 					id		path	string true "sub merchant id"
// @Param 					type	path	string true "Oneof merchant or merchant-platform"
// @Success					200	{object}	response.ApiResponse{data=disbursementModel.DailyTransactionLimitResponse}
// @Failure					500 {object}	response.ApiResponse
// @Router					/api/v1/sub-merchants/{id}/daily-limits/{type} [get]
// @Security				Bearer
func (h *SubMerchantController) GetSubMerchantDailyLimit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/api/v1/subMerchant/GetSubMerchantDailyLimit")
	defer segment.End()

	_, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	subMerchantID := chi.URLParam(r, "id")
	_, err := uuid.Parse(subMerchantID)
	if err != nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, constant.ErrMerchantIDNotValid))
		return
	}

	merchantType := r.PathValue("type")
	if err := h.validate.VarCtx(ctx, merchantType, "oneof=merchant merchant-platform"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, errors.New("merchant type not registered")))
		return
	}

	resp, err := h.disbursementSvc.GetDailyTransactionLimit(ctx, subMerchantID, merchantType)
	if errors.Is(err, constant.ErrForbiddenAccess) {
		w.WriteHeader(http.StatusNoContent) // Note: When sub-merchants access it then response is code 204 No Content
		return
	}

	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, resp)
}
