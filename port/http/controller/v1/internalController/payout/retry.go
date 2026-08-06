package internalPayoutController

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// RetryBulk		godoc
// @Summary			Retry bulk disbursement
// @Description		Retry bulk disbursement
// @ID				internal-retry-bulk-disbursements
// @Tags			Internal - Disbursement
// @Accept			json
// @Produce			json
// @Param			id		query		string					true	"id of bulk disbursement"
// @Success			200  	{object}	response.ApiResponse
// @Failure			500  	{object}	response.ApiResponse
// @Router			/internal/v1/payouts/:id/retry [post]
// @Security		Bearer
func (c *InternalPayoutController) RetryBulk(w http.ResponseWriter, r *http.Request) {
	var (
		merchantID    string
		err           error
		payoutsObject []disbursementModel.PayoutObjectForRetry
		ctx                 = r.Context()
		page          int64 = 1
		perPage       int64 = constant.DefaultPaginationPageSize
	)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/internalController/payout/RetryBulk")
	defer segment.End()

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, fmt.Errorf("disbursement id is required")))
		return
	}

	// Merchant info from JWT
	merchantInfoFromCtx := r.Context().Value(constant.CtxMerchantInfo)
	merchantCtx, ok := merchantInfoFromCtx.(*merchantModel.MerchantAuthTokenClaims)
	if !ok {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound))
		return
	}

	merchantID = merchantCtx.MerchantId
	httputil.BindSubmerchantID(r, &merchantID)

	// set filter
	filter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID: merchantID,
		BulkID:     id,
	}

	bulkDisbursementRes, err := c.disbursementSvc.GetBulkDisbursementForOpenApiByID(ctx, filter, page, perPage)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	if bulkDisbursementRes == nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrors.New(response.HttpErrNotFound, constant.ErrBulkDisbursementNotFound))
		return
	}

	// Set model
	requestPayload := disbursementModel.RetryBulkRequest{
		BulkDisbursementID: id,
		MerchantID:         merchantID,
	}

	err = c.disbursementSvc.RetryBulk(r.Context(), &requestPayload)
	if err != nil {
		response.SendOpenApiNonSnapResponseError(ctx, w, err)
		return
	}

	bulkDisbursement := bulkDisbursementRes.Data.(*disbursementModel.GetBulkDisbursementForOpenApiByIDResponse)

	// build payoutObjectsForRetry
	for _, payout := range bulkDisbursement.Payouts {
		payoutObject := disbursementModel.PayoutObjectForRetry{
			ReferenceID: payout.ReferenceID,
			InquiryID:   payout.InquiryID,
			ChannelCode: payout.ChannelCode,
			ChannelInformation: disbursementModel.PayoutChannelInformation{
				AccountNumber: payout.ChannelInformation.AccountNumber,
				AccountName:   payout.ChannelInformation.AccountName,
			},
			Amount: commonModel.Amount{
				Currency: payout.Amount.Currency,
				Value:    payout.Amount.Value,
			},
			Description: payout.Description,
		}
		payoutsObject = append(payoutsObject, payoutObject)
	}

	// build response
	payoutResponse := disbursementModel.RetryDisbursementFromOpenApiResponse{
		UUID:       bulkDisbursement.UUID,
		MerchantID: merchantID,
		Payouts:    payoutsObject,
		Status:     bulkDisbursement.Status,
		CreatedAt:  bulkDisbursement.CreatedAt,
		UpdatedAt:  bulkDisbursement.UpdatedAt,
	}

	// publish activity, do nothing on error
	_ = c.rabbitMqExt.PublishActivity(
		ctx,
		&merchantCtx.MerchantId,
		nil,
		constant.TagDisbursement,
		constant.ActivityMerchantRetryDisbursement,
		payoutResponse,
	)

	response.SendOpenApiResponseOK(w, payoutResponse)
}
