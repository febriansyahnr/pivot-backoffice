package merchant

import (
	"encoding/json"
	"errors"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"net/http"
)

// Assign		godoc
// @Summary		Assign user to merchant endpoint
// @Description	Assign user to merchant endpoint
// @ID			crm-merchant-assign-user
// @Tags		CRM - Merchant
// @Accept		json
// @Produce		json
// @Param		Request	body		merchant.MerchantAssignRequest true "JSON Body for assign User to Merchant"
// @Success		200  	{object}	response.Response{data=merchant.MerchantAssignResponse}
// @Failure		500  	{object}	response.Response
// @Router		/api/v1/merchants/assign [post]
// @Security	Bearer
func (c *CRMMerchantController) Assign(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/merchant/Assign")
	defer segment.End()

	var (
		err  error
		user *userModel.User
	)

	var payload merchantModel.MerchantAssignRequest
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	if err = c.validate.Struct(payload); err != nil {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	user, err = c.userSvc.FindUserByID(ctx, payload.UserID)
	if err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// check if user already assigned to merchant
	if user.MerchantId != "" {
		response.SendGeneralResponseError(w, pkgError.New(response.HttpErrConflict, errors.New("user already assigned to merchant")))
		return
	}

	// assign user to merchant
	user.MerchantId = payload.MerchantID
	if err = c.userSvc.Update(ctx, user); err != nil {
		response.SendGeneralResponseError(w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&payload.MerchantID,
		&user.ID,
		constant.TagMerchant,
		constant.ActivityUserAssignMerchant,
		payload,
	)

	response.SendGeneralResponseOK(w, "User assigned to merchant successfully")
}
