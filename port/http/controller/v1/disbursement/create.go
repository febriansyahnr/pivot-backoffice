package disbursementController

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	beneficiaryModel "github.com/paper-indonesia/pivot-backoffice/internal/model/beneficiaryAccount"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/monitor"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/shopspring/decimal"
)

// CreateSingle		godoc
// @Summary			Create single disbursement
// @Description		Create single disbursement
// @ID				api-disbursement-create-single
// @Tags			API - Disbursement
// @Accept			json
// @Produce			json
// @Param			Request	body		disbursementModel.CreateSingleRequest true "JSON Body for Create Single Disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/api/v1/disbursements/single/create [post]
// @Security		Bearer
func (c *Controller) CreateSingle(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/CreateSingle")
	defer segment.End()

	var (
		err error
		now = time.Now()
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	var requestPayload disbursementModel.CreateSingleRequest
	if err = json.NewDecoder(r.Body).Decode(&requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// Validate request
	if err = c.validate.Struct(requestPayload); err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	_, err = util.ValidateMagicNumber(http.MethodPost, requestPayload.BeneficiaryAccountNo)
	if c.config.Environment != constant.EnvironmentProduction && err != nil {
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	ctx, configs, err := c.merchant.GetMerchantIdForConfigs(ctx, user.MerchantId, true)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}
	merchant, _ := ctx.Value(constant.CtxMerchantData).(*merchant.Merchant)

	trxConfig, err := c.disbursementSvc.GetTransactionConfig(ctx, configs.MerchantTransactionConfig)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	maxAmount := trxConfig.MaxAmount
	if c.disbursementSvc.IsBankcodeOverbookingChannelAllowed(ctx, requestPayload.BeneficiaryBankCode, requestPayload.MerchantID) {
		maxAmount = c.config.DisbursementConfig.OverbookingBankMaxAmount

		beneficiaryAccount, errBeneficiary := c.beneficiaryAccountSvc.FindByBankCodeAndAccountNo(ctx, &beneficiaryModel.CheckAccountRequest{
			BeneficiaryAccountNo: requestPayload.BeneficiaryAccountNo,
			BeneficiaryBankCode:  requestPayload.BeneficiaryBankCode,
			MerchantID:           user.MerchantId,
			AdditionalInfo:       map[string]any{},
		})
		if errBeneficiary != nil {
			response.SendApiResponseError(ctx, w, errBeneficiary)
			return
		}
		if max, isValid := c.disbursementSvc.IsMerchantAllowedExcludeBeneficiaryRules(ctx, user.MerchantId, requestPayload.Amount.InexactFloat64()); isValid {
			maxAmount = max
		} else if c.disbursementSvc.IsMerchantAllowedToUseBeneficiaryCustomRule(ctx, user.MerchantId, beneficiaryAccount.MetadataObj.BeneficiaryPayoutLimitRule != nil) {
			maxAmount = c.config.DisbursementConfig.OverbookingBankMaxAmountForCustomRule
		}
	}

	if requestPayload.Amount.LessThan(decimal.NewFromFloat(trxConfig.MinAmount)) {
		response.SendApiResponseError(
			ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf(minAmountErrFmt, util.ConvertFloatToCurrency(trxConfig.MinAmount))),
		)
		return

	} else if requestPayload.Amount.GreaterThan(decimal.NewFromFloat(maxAmount)) {
		response.SendApiResponseError(
			ctx, w, errors.New(response.HttpErrRequest, fmt.Errorf(maxAmountErrFmt, util.ConvertFloatToCurrency(maxAmount))),
		)
		return
	}

	// Validate referenceID
	isExist := c.disbursementSvc.IsExistReferenceID(ctx, user.MerchantId, requestPayload.ReferenceID)
	if isExist {
		err = constant.ErrDisbursementReferenceIdAlreadyExist
		response.SendApiResponseError(ctx, w, errors.New(response.HttpErrRequest, err))
		return
	}

	// set additional request
	requestPayload.MerchantID = merchant.UUID
	requestPayload.MerchantName = merchant.Name
	requestPayload.CreatedBy = &user.UUID
	requestPayload.CreatedFrom = constant.DisbursementCreatedFromMerchantPortal

	ww := monitor.WrapResponse(w, r)
	defer func() {
		monitor.WriteAndSend(
			ctx, "disbursement-create-single", now, ww, err, func() []string {
				return []string{
					fmt.Sprintf("merchant_id:%s", requestPayload.MerchantID),
					fmt.Sprintf("beneficiary_bank_name:%s", requestPayload.BeneficiaryBankName),
					fmt.Sprintf("amount:%s", requestPayload.Amount),
				}
			},
		)
	}()

	data, err := c.disbursementSvc.CreateSingle(ctx, &requestPayload)
	if err != nil {
		if typ, _ := errors.ExtractError(err); typ == response.HttpErrRequest {
			response.SendApiResponseError(ctx, ww, err)
			return
		}
		response.SendApiResponseError(ctx, ww, errors.New(response.HttpErrInternal, err))
		return
	}
	response.SendApiResponseOK(ww, data)
}
