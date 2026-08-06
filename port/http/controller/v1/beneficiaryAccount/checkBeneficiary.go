package beneficiaryAccountController

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/shopspring/decimal"
)

// CheckBeneficiary	godoc
// @Summary			Check beneficiary account
// @Description		Check beneficiary account
// @ID				api-beneficiary-account-check-account
// @Tags			API - Beneficiary Account
// @Accept			json
// @Produce			json
// @Param			Request	body		beneficiaryAccountModel.CheckAccountRequest true "JSON Body for Check Beneficiary Account"
// @Success			200  	{object}	response.ApiResponse{data=beneficiaryAccountModel.CheckAccountResponse}
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/beneficiary-accounts/inquiry [post]
// @Security		Bearer
func (c *Controller) CheckBeneficiary(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/BeneficiaryAccount/CheckBeneficiary")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var payload beneficiaryAccountModel.CheckAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Trim spaces
	payload.BeneficiaryAccountNo = strings.ReplaceAll(payload.BeneficiaryAccountNo, " ", "")

	if err := c.validate.Struct(payload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	payload.MerchantID = user.MerchantId
	payload.AdditionalInfo = map[string]any{}
	account, err := c.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &payload)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	if constant.IsAccountInquiryVirtualAccountFlagDisplayedForMerchant(payload.MerchantID) {
		account.AdditionalInfo = &beneficiaryAccountModel.AccountAdditionalInfo{
			IsVirtualAccount: account.MetadataObj.IsVirtualAccount,
		}
	}

	trxConfig, err := c.disbursementSvc.GetTransactionConfig(ctx, payload.MerchantID)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	account.MetadataObj.IsOverbooking = false
	account.MetadataObj.MaxAmount = decimal.NewFromFloat(trxConfig.MaxAmount)
	if c.disbursementSvc.IsBankcodeOverbookingChannelAllowed(ctx, account.BeneficiaryBankCode, account.MerchantID) {
		account.MetadataObj.IsOverbooking = true
		account.MetadataObj.MaxAmount = decimal.NewFromFloat(c.config.DisbursementConfig.OverbookingBankMaxAmount)

		if max, isValid := c.disbursementSvc.IsMerchantAllowedExcludeBeneficiaryRules(ctx, payload.MerchantID, 0); isValid {
			// Set maximum amount to a reasonable business limit instead of math.MaxFloat64
			account.MetadataObj.MaxAmount = decimal.NewFromFloat(max)
		} else if c.disbursementSvc.IsMerchantAllowedToUseBeneficiaryCustomRule(ctx, payload.MerchantID, account.MetadataObj.BeneficiaryPayoutLimitRule != nil) {
			account.MetadataObj.MaxAmount = decimal.NewFromFloat(c.config.DisbursementConfig.OverbookingBankMaxAmountForCustomRule)
		}
	}

	ctx, configs, err := c.merchantSvc.GetMerchantIdForConfigs(ctx, payload.MerchantID, false)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	feeRequest := feeModel.GetPayoutTrxFeeRequest{
		MerchantId:   payload.MerchantID,
		MerchantType: configs.MerchantType,
		BankCode:     payload.BeneficiaryBankCode,
	}
	feeRequest.ParentMerchantId, _ = ctx.Value(constant.CtxParentMerchantId).(string)

	feeResult, err := c.feeSvc.GetPayoutTransactionFee(ctx, feeRequest)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	account.MetadataObj.PayoutFeeAmount = feeResult.ToFeeResponse().TotalAmount

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&user.MerchantId,
		&user.UUID,
		constant.TagDisbursement,
		constant.ActivityUserCheckBeneficiaryAccount,
		payload,
	)

	if account.BeneficiaryAccountName == "" {
		response.SendApiResponseSuccess(w, http.StatusAccepted, constant.ErrRequestInProgress.Error(), account)
		return
	}

	response.SendApiResponseOK(w, account)
}
