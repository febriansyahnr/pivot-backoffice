package merchant

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// TransactionConfig	godoc
// @Summary				Update Merchant / Sub Merhcnat KYC Status.
// @Description			this endpoint can change the kyc information of the merchant
// @ID					crm-merchant-update-kyc
// @Tags				CRM - Merchant
// @Accept				json
// @Produce				json
// @Param 				id		path		string true "Merchant Id or Sub-Merchant Id"
// @Param				Request	body		merchant.UpdateMerchantKYCRequest true "JSON Body for update merchant kyc"
// @Success				200  	{object}	response.Response{data=merchant.UpdateMerchantKYCResponse}
// @Failure				500  	{object}	response.Response
// @Router				/crm/v1/merchants/{id}/kyc [put]
// @Header       		all     {string}  X-CRM-Key "{"key": "value"}"
func (c *CRMMerchantController) UpdateKYC(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/UpdateKYC")
	defer segment.End()

	request := merchant.UpdateMerchantKYCRequest{
		MerchantID: r.PathValue("id"),
	}

	if err := uuid.Validate(request.MerchantID); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, errors.New("invalid merchant id")))
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(request); err != nil {
		response.SendGeneralResponseError(w, pkgErrs.New(response.HttpErrRequest, err))
		return
	}

	merchant, err := c.merchantSvc.UpdateKYC(ctx, request)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&merchant.UUID,
		util.ValueToPtr(constant.UserSystemType),
		constant.TagMerchant,
		constant.ActivityUserChangeKYCInfo,
		request,
	)

	response.SendGeneralResponseOK(w, merchant)
}
