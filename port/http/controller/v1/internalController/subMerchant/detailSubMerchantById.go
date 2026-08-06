package submerchant

import (
	"context"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
)

func (c *SubMerchantInternalController) DetailSubMerchantByID(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/openApi/v1/subMerchant/DetailSubMerchantByID")
	defer segment.End()

	var (
		err         error
		subMerchant *merchantModel.Merchant
	)

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))

	merchantInfo := ctx.Value(constant.CtxMerchantInfo)
	merchant, ok := merchantInfo.(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, err))
		return
	}

	loggedInMerchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &loggedInMerchantId)
	loggedInUserType := constant.UserTypeMerchant
	httputil.BindLoggedInUserType(r, &loggedInUserType)

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, constant.ErrRequiredSubmerchantId))
		return
	}

	if subMerchant, err = c.merchantSvc.FindMerchantByID(ctx, id); err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrDatabase, err))
		return

	} else if subMerchant == nil || subMerchant.ParentID.String != merchant.MerchantId {
		// Treat parent mismatch as "not found" for consistency and to avoid leaking existence
		// of sub-merchants owned by other parents.
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrResourceNotFound("sub-account", id))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrNotFound, constant.ErrSubMerchantNotFound))
		return
	}

	response.SendOpenApiResponseOK(w, subMerchant.ToSubMerchantResponse())
}
