package submerchant

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *SubMerchantInternalController) Create(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/subMerchant/Create")
	defer segment.End()

	ctx = context.WithValue(ctx, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendGeneralResponseError))

	merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// merchant.MerchantId comes from the authenticated JWT claim, so a lookup miss here
	// is a system/infra error (DB failure) rather than a bad request.
	parentMerchant, err := c.merchantSvc.FindMerchantByID(ctx, merchant.MerchantId)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrDatabase, err))
		return
	}

	loggedInMerchantId := merchant.MerchantId
	httputil.BindSubmerchantID(r, &loggedInMerchantId)
	loggedInUserType := constant.UserTypeMerchant
	httputil.BindLoggedInUserType(r, &loggedInUserType)

	// Decode request into MerchantRequest
	var payload merchantModel.MerchantRequest
	payload.RequesterID = loggedInMerchantId
	payload.RequesterType = loggedInUserType
	payload.ParentID = loggedInMerchantId
	payload.MerchantStatus = constant.MerchantStatusActive // Set automatically to active

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrInvalidPayload(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}
	payload.KYCStatus = payload.SubAccountType

	if payload.BankAccount == nil {
		// Set a default value when the field is not specified because the struct is shared with the create merchant process,
		// which does not require a bank account (added separately).
		// This results in an invalid request as bank account details are not provided.
		payload.BankAccount = &merchantModel.MerchantBankAccountRequest{}
	}

	// Set enforcement flag based on feature flag
	payload.EnforceMandatoryFields = constant.ShouldEnforceStandardizedAddress(parentMerchant.UUID, parentMerchant.CreatedAt)

	if err := c.validate.Struct(payload); err != nil {
		ctx = context.WithValue(ctx, constant.CtxErrorInfo, constant.NewErrFieldValidation(err))
		response.SendOpenApiNonSnapResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	if payload.KYCStatus == constant.MerchantKYCTypeKYC {
		payload.MerchantStatus = constant.MerchantStatusCreated
		payload.KYCStatus = constant.KYCStatusWaitingForDocument
	} else {
		payload.KYCStatus = constant.KYCStatusNotRequired
	}

	data, err := c.merchantSvc.CreateSubMerchant(ctx, &payload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}
	result := data.ToSubMerchantResponse()
	result.AutoWithdrawal = payload.AutoWithdrawal
	result.SubAccountType = payload.SubAccountType
	if result.SubAccountType == "" {
		result.SubAccountType = constant.MerchantKYCTypeNonKYC
	}
	result.SubAccountKycStatus = payload.KYCStatus

	response.SendOpenApiResponseOK(w, result)
}
