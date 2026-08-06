package disbursementController

import (
	"errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	feeModel "github.com/paper-indonesia/pivot-backoffice/internal/model/fee"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

// GetTransactionConfig	godoc
// @Summary				Endpoint to get disbursement configuration details
// @Description			Endpoint to get disbursement configuration details
// @ID					api-disbursement-transaction-config
// @Tags				API - Disbursement
// @Accept				json
// @Produce				json
// @Success				200	{object}	response.ApiResponse{data=disbursementModel.TransactionConfigResp}
// @Failure				500 {object}	response.ApiResponse
// @Router				/api/v1/disbursements/configs [get]
// @Security			Bearer
func (h *Controller) GetTransactionConfig(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetTransactionConfig")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	ctx, configs, err := h.merchant.GetMerchantIdForConfigs(ctx, user.MerchantId, false)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	transaction, err := h.disbursementSvc.GetTransactionConfig(ctx, configs.MerchantTransactionConfig)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	var result feeModel.FeeResponseder

	if configs.MerchantType == constant.MerchantTypeSubMerchant {
		parentMerchantId, _ := ctx.Value(constant.CtxParentMerchantId).(string)

		result, err = h.feeSvc.GetTransactionFeeOnBehalf(
			ctx, &feeModel.GetTrxFeeOnBehalfRequest{
				MerchantId:    parentMerchantId,
				SubMerchantId: user.MerchantId,
				Reference:     constant.ReferenceDisbursement,
			},
		)
	} else {
		_, result, err = h.feeSvc.GetFeeCalculationAndDetail(ctx, &feeModel.GetFeeRequest{
			MerchantID: user.MerchantId,
			Reference:  constant.ReferenceDisbursement,
		})
	}
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, disbursementModel.TransactionConfigResp{
		TransactionConfig: transaction,
		FeeDetail:         result.ToFeeResponse(),
	})
}

// GetDailyTransactionLimit	godoc
// @Summary					Endpoint to view daily transaction limit, total transactions (amount), and remaining allocation.
// @Description				Endpoint to view daily transaction limit, total transactions (amount), and remaining allocation.
// @ID						api-disbursement-get-daily-transaction-limit
// @Tags					API - Disbursement
// @Accept					json
// @Produce					json
// @Param 					type	path	string true "Oneof merchant or merchant-platform"
// @Success					200	{object}	response.ApiResponse{data=disbursementModel.DailyTransactionLimitResponse}
// @Failure					500 {object}	response.ApiResponse
// @Router					/api/v1/disbursements/daily-limits/{type} [get]
// @Security				Bearer
func (h *Controller) GetDailyTransactionLimit(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetTransactionConfig")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	merchantType := r.PathValue("type")
	if err := h.validate.VarCtx(ctx, merchantType, "oneof=merchant merchant-platform"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, errors.New("merchant type not registered")))
		return
	}

	if resp, err := h.disbursementSvc.GetDailyTransactionLimit(ctx, user.MerchantId, merchantType); err != nil {
		if errors.Is(err, constant.ErrForbiddenAccess) {
			w.WriteHeader(http.StatusNoContent) // Note: When sub-merchants access it then response is code 204 No Content

		} else {
			response.SendApiResponseError(ctx, w, err)
		}
	} else {
		response.SendApiResponseOK(w, resp)
	}
}

// GetTransactionLimit	godoc
// @Summary				Endpoint to view transaction limits (min & max) and daily transaction limits (daily accumulation).
// @Description			Endpoint to view transaction limits (min & max) and daily transaction limits (daily accumulation)
// @ID					api-disbursement-get-transaction-limit
// @Tags				API - Disbursement
// @Accept				json
// @Produce				json
// @Success				200	{object}	response.ApiResponse{data=disbursementModel.DailyTransactionLimitResponse}
// @Failure				500 {object}	response.ApiResponse
// @Router				/api/v1/disbursements/limits [get]
// @Security			Bearer
func (c *Controller) GetTransactionLimit(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetTransactionLimit")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	if resp, err := c.merchant.GetTransactionConfig(ctx, user.MerchantId); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}

// GetTransactionLimitSubMerchant	godoc
// @Summary							Endpoint to view transaction limits (min & max) and daily transaction limits (daily accumulation) for sub-merchant.
// @Description						Endpoint to view transaction limits (min & max) and daily transaction limits (daily accumulation) for sub-merchant.
// @ID								api-disbursement-get-transaction-limit-for-sub-merchant
// @Tags							API - Disbursement
// @Accept							json
// @Produce							json
// @Success							200	{object}	response.ApiResponse{data=disbursementModel.DailyTransactionLimitResponse}
// @Failure							500 {object}	response.ApiResponse
// @Router							/api/v1/disbursements/limits/sub-merchants/{id} [get]
// @Security						Bearer
func (c *Controller) GetTransactionLimitSubMerchant(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/disbursement/GetTransactionLimitSubMerchant")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	subMerchantId := r.PathValue("id")
	if err := uuid.Validate(subMerchantId); err != nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}

	subMerchant, err := c.merchant.FindMerchantByID(ctx, subMerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return

	} else if subMerchant == nil {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrMerchantNotFound))
		return

	} else if subMerchant.ParentID.String != user.MerchantId {
		response.SendApiResponseError(ctx, w, pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess))
		return
	}

	if resp, err := c.merchant.GetTransactionConfig(ctx, subMerchant.UUID); err != nil {
		response.SendApiResponseError(ctx, w, err)

	} else {
		response.SendApiResponseOK(w, resp)
	}
}
