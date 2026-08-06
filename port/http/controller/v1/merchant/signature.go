package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GenOpenAPISignature	godoc
// @Summary				Utility endpoint to generate b2b signature (X-Signature)
// @Description			Utility endpoint to generate b2b signature (X-Signature)
// @ID					merchant-utils-gen-b2b-signature
// @Tags				API - Merchant
// @Accept				json
// @Produce				json
// @Param				Request	body merchant.GenSignatureReq true "JSON Body for Generate Open API signature"
// @Success				200  	{object}	response.ApiResponse{data=merchant.GenSignatureResp}
// @Failure				500  	{object}	response.ApiResponse
// @Router				/api/v1/merchants/utilities/generate-signature [post]
// @Security			Bearer
func (h *MerchantController) GenOpenAPISignature(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/GenOpenAPISignature")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	req := merchant.GenSignatureReq{
		MerchantId: user.MerchantId,
		Timestamp:  r.Header.Get(constant.HeaderXTimestamp),
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := h.validate.Struct(&req); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if signature, err := h.merchantSvc.GenOpenAPISignature(ctx, &req); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, merchant.GenSignatureResp{Signature: signature})
	}

}
