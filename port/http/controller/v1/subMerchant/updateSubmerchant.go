package subMerchant

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantController) Update(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/api/v1/subMerchant/Update")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	id := chi.URLParam(r, "id")
	if err := uuid.Validate(id); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf("sub merchant id is required")))
		return
	}

	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantAuth.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	var payload merchantModel.UpdateMerchantRequest
	payload.RequesterID = merchantAuth.MerchantId
	payload.RequesterType = constant.UserTypeMerchant
	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	SetDefaultMerchantValuesUpdate(&payload, merchant)

	payload.ID = id
	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	_, err = c.merchantSvc.UpdateSubMerchant(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendGeneralResponseOK(w, "Update submerchant success")
}

func SetDefaultMerchantValuesUpdate(payload *merchantModel.UpdateMerchantRequest, merchant *merchantModel.Merchant) {
	if payload.Logo == "" {
		payload.Logo = merchant.Logo
	}
	if payload.MerchantEmail == "" {
		payload.MerchantEmail = merchant.MerchantEmail
	}
	if payload.MerchantPhone == "" {
		payload.MerchantPhone = merchant.MerchantPhone
	}
	if payload.PICJobTitle == "" {
		payload.PICJobTitle = merchant.PICJobTitle.String
	}
	if payload.PICPhone == "" {
		payload.PICPhone = merchant.PICPhone
	}
	if payload.Address == "" {
		if merchant.Address != "" {
			payload.Address = merchant.Address
		} else {
			payload.Address = "-"
		}
	}
	if payload.PostCode == "" {
		if merchant.PostCode != "" {
			payload.PostCode = merchant.PostCode
		} else {
			payload.PostCode = "0"
		}
	}
	if payload.Status == "" {
		payload.Status = merchant.Status
	}
}
