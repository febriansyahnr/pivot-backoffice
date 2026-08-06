package merchant

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FindByMerchantID		godoc
// @Summary				Get merchant active product
// @Description			harysa has several products that can be accessed by merchant, but it should confirm into harsya team
// @ID					merchant-active-product
// @Tags				API - Merchant
// @Accept				json
// @Produce				json
// @Success				200  		{object}	response.ApiResponse{data=[]merchant.MerchantResponse}
// @Failure				500  		{object}	response.ApiResponse
// @Router				/api/v1/merchants/actived-products [get]
// @Security			Bearer
func (c *MerchantController) GetActiveProducts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/merchant/GetActiveProducts")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	products, err := c.productSvc.GetMerchantActiveProducts(r.Context(), user.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, products)
}
