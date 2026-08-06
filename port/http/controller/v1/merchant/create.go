package merchant

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgError "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// Create		godoc
// @Summary		Create merchant endpoint
// @Description	Create merchant endpoint
// @ID			merchant-create
// @Tags		API - Merchant
// @Accept		json
// @Produce		json
// @Param		Request	body		merchant.MerchantRequest true "JSON Body for Create Merchant"
// @Success		200  	{object}	response.ApiResponse{data=merchant.MerchantResponse}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/merchants/create [post]
// @Security	Bearer
func (c *MerchantController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/merchant/Create")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// check if user already have merchant
	if user.MerchantId != "" {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrForbidden, errors.New("user already have merchant")))
		return
	}

	var payload merchantModel.MerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, pkgError.New(response.HttpErrRequest, err))
		return
	}

	merchantData := &merchantModel.Merchant{
		UUID:          uuid.New().String(),
		ExternalId:    util.GenerateULID(),
		Name:          payload.Name,
		ShortName:     payload.ShortName,
		Description:   payload.Description,
		Address:       payload.Address,
		DistrictId:    payload.DistrictId,
		PostCode:      payload.PostCode,
		Logo:          payload.Logo,
		MerchantEmail: payload.MerchantEmail,
		MerchantPhone: payload.MerchantPhone,
		PICEmail:      payload.PICEmail,
		PICPhone:      payload.PICPhone,
		Status:        constant.MerchantStatusActive,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	transactionConfig := map[string]interface{}{
		"autoWithdrawal": constant.AutoWithdrawalStateOFF,
	}

	if c.config.WithdrawalConfig.AutoWithdrawalDefaultState == constant.AutoWithdrawalStateON {
		transactionConfig["autoWithdrawal"] = constant.AutoWithdrawalStateON
	}

	if payload.AutoWithdrawal != nil {
		transactionConfig["autoWithdrawal"] = *payload.AutoWithdrawal
	}
	merchantData.TransactionConfigs.Valid = true
	merchantData.TransactionConfigs.JSONText, _ = json.Marshal(transactionConfig)

	if err := c.merchantSvc.Create(ctx, merchantData, payload.UserID); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagMerchant,
		constant.ActivityUserCreateMerchant,
		payload,
	)

	response.SendApiResponseOK(w, merchantData.ToResponse())
}
