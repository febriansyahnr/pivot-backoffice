package subMerchant

import (
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/api/v1/subMerchantCreate")
	defer segment.End()

	merchantAuth, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchant, err := c.merchantSvc.FindMerchantByID(ctx, merchantAuth.MerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	loggedInMerchantId := merchantAuth.MerchantId
	httputil.BindSubmerchantID(r, &loggedInMerchantId)
	loggedInUserType := constant.UserTypeMerchant
	httputil.BindLoggedInUserType(r, &loggedInUserType)

	var payload merchantModel.CreateSubMerchantRequest
	payload.RequesterID = loggedInMerchantId
	payload.UserID = merchantAuth.UUID
	payload.RequesterType = loggedInUserType
	payload.ParentID = loggedInMerchantId
	payload.MerchantStatus = constant.MerchantStatusActive // Set automatically to active

	if err = json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Set defaults here before validation
	SetDefaultMerchantValuesCreate(&payload, merchant)

	if payload.BankAccount == nil {
		// Set a default value when the field is not specified because the struct is shared with the create merchant process,
		// which does not require a bank account (added separately).
		// This results in an invalid request as bank account details are not provided.
		payload.BankAccount = &merchantModel.MerchantBankAccountRequest{}
	}
	if err = c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if payload.KYCStatus == constant.MerchantKYCTypeNonKYC {
		payload.KYCStatus = constant.KYCStatusNotRequired
	} else if payload.KYCStatus == constant.MerchantKYCTypeKYC {
		payload.KYCStatus = constant.KYCStatusWaitingForDocument
	}

	data, err := c.merchantSvc.CreateSubMerchant(ctx, &merchantModel.MerchantRequest{
		Name:              payload.Name,
		MerchantStatus:    payload.MerchantStatus,
		ShortName:         payload.ShortName,
		Description:       payload.Description,
		Website:           payload.Website,
		Address:           payload.Address,
		DistrictId:        payload.DistrictId,
		PostCode:          payload.PostCode,
		Logo:              payload.Logo,
		MerchantEmail:     payload.MerchantEmail,
		MerchantPhone:     payload.MerchantPhone,
		BusinessType:      payload.BusinessType,
		BusinessStructure: payload.BusinessStructure,
		BusinessCountry:   payload.BusinessCountry,
		ParentIndustry:    payload.ParentIndustry,
		ChildIndustry:     payload.ChildIndustry,
		MCC:               payload.MCC,
		CountryOfEntity:   payload.CountryOfEntity,
		DigitalStatus:     payload.DigitalStatus,
		PICInvitation:     payload.PICInvitation,
		PICName:           payload.PICName,
		PICEmail:          payload.PICEmail,
		PICPhone:          payload.PICPhone,
		PICJobTitle:       payload.PICJobTitle,
		AutoWithdrawal:    payload.AutoWithdrawal,
		KYCStatus:         payload.KYCStatus,
		ParentID:          payload.ParentID,
		RequesterID:       payload.RequesterID,
		RequesterType:     payload.RequesterType,
		UserID:            payload.UserID,
		BankAccount:       payload.BankAccount,
	})
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	result := data.ToSubMerchantResponse()
	result.AutoWithdrawal = payload.AutoWithdrawal

	response.SendGeneralResponseOK(w, result)
}

func SetDefaultMerchantValuesCreate(payload *merchantModel.CreateSubMerchantRequest, merchant *merchantModel.Merchant) {
	if payload.Logo == "" {
		payload.Logo = merchant.Logo
	}
	if payload.MerchantStatus == "" {
		payload.MerchantStatus = constant.MerchantStatusActive
	}
	if payload.Website == "" {
		payload.Website = merchant.Website
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
}
