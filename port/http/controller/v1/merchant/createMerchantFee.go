package merchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// CreateMerchantFee	godoc
// @Summary				Create merchant fee endpoint
// @Description			Create merchant fee endpoint
// @ID					merchant-fee-create
// @Tags				API - Merchant Fee
// @Accept				json
// @Produce				json
// @Param				Request	body		merchant.MerchantRequest true "JSON Body for Create Merchant fee"
// @Success				200  	{object}	response.Response{data=merchant.MerchantFeeResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/fee [post]
// @Security			Bearer
func (c *MerchantController) CreateMerchantFee(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/CreateMerchantFee")
	defer segment.End()

	var payload merchantModel.NewMerchantFeeRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, errors.New(response.HttpErrRequest, err))
		return
	}

	resp, err := c.merchantSvc.CreateMerchantFee(ctx, &payload)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		nil,
		util.ValueToPtr("retool"),
		constant.TagMerchant,
		constant.ActivityUserCreateMerchantFee,
		payload,
	)

	response.SendGeneralResponseOK(w, resp)
}
