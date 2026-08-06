package merchant

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// FindMerchantFeeByMerchantIDAndType		godoc
// @Summary									Find merchant fee by merchant id and type
// @Description								Find merchant fee by merchant id and type
// @ID										merchant-fee-find-by-merchant-id-and-type
// @Tags									API - Merchant Fee
// @Accept									json
// @Produce									json
// @Param       							type    	query     	string  false  "param for fee type"
// @Success									200  		{object}	response.ApiResponse{data=merchant.MerchantFeeResponse}
// @Failure									500  		{object}	response.ApiResponse
// @Router									/api/v1/merchants/fee [get]
// @Security								Bearer
func (c *MerchantController) FindMerchantFeeByMerchantIDAndType(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/FindMerchantFeeByMerchantIDAndType")
	defer segment.End()

	var (
		err        error
		defaultFee = constant.TypeDisbursement
	)

	feeType := r.URL.Query().Get("type")
	if feeType == "" {
		feeType = defaultFee
	}

	feeMethod := r.URL.Query().Get("method")

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	feeAmount, feeDetail, err := c.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
		MerchantID:    user.MerchantId,
		Reference:     feeType,
		PaymentMethod: feeMethod,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, feeDetail.WithTotalAmountResponse(feeAmount))
}
