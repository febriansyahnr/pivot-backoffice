package xbPayoutController

import (
	"net/http"
	"sort"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	statusHistoriesModel "github.com/paper-indonesia/pivot-backoffice/internal/model/statusHistory"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func (c *xbPayoutController) GetDetail(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetDetail")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// get id from url path
	id := chi.URLParam(r, "id")
	if errId := uuid.Validate(id); errId != nil {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdIsRequired))
		return
	}

	disbursement, err := c.disbursementSvc.FindByID(ctx, id)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	// check disbursement with merchant
	if disbursement.MerchantID != user.MerchantId {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrMerchantIsNotMatch))
		return
	}

	response.SendApiResponseOK(w, buildDetailResponse(disbursement))
}

func buildDetailResponse(disbursement *disbursementModel.DisbursementWithTransaction) (resp *xbModel.GetXbPayoutDetailResponse) {
	fee := decimal.NewFromFloat(disbursement.MetadataObj.FeeDetail.FinalAmount)

	sourceAmount := disbursement.MetadataObj.XbDetail.SourceAmount
	totalAmount := disbursement.MetadataObj.XbDetail.TotalAmount

	status := ""
	if disbursement.ReasonType != nil {
		status = *disbursement.ReasonType
	}
	statusDescription := ""
	if disbursement.ReasonDescription != nil {
		statusDescription = *disbursement.ReasonDescription
	}
	remark := ""
	if disbursement.Remark != nil {
		remark = *disbursement.Remark
	}

	response := xbModel.GetXbPayoutDetailResponse{
		UUID:                disbursement.UUID,
		MerchantId:          disbursement.MerchantID,
		ReferenceId:         disbursement.ReferenceID,
		SourceCurrency:      disbursement.MetadataObj.XbDetail.SourceCurrency,
		DestinationCurrency: disbursement.Currency,
		DestinationAmount:   disbursement.Amount,
		FxRate:              disbursement.MetadataObj.XbDetail.FxRate,
		DestinationFxRate:   disbursement.MetadataObj.XbDetail.DestinationFxRate,
		SourceAmount:        sourceAmount,
		Fee:                 fee,
		TotalAmount:         totalAmount,
		CreatedAt:           disbursement.CreatedAt,
		UpdatedAt:           disbursement.UpdatedAt,
		ExpiredAt:           disbursement.MetadataObj.XbDetail.ExpiredAt,
		Status:              status,
		StatusDescription:   statusDescription,
		PurposeCode:         disbursement.MetadataObj.XbDetail.PurposeCode,
		Remark:              remark,
		SenderData:          disbursement.MetadataObj.XbDetail.SenderData,
		BeneficiaryId:       disbursement.MetadataObj.XbDetail.BeneficiaryId,
		BeneficiaryData:     disbursement.MetadataObj.XbDetail.BeneficiaryData,
		RoutingCode:         disbursement.MetadataObj.XbDetail.RoutingCode,
		RoutingValue:        disbursement.MetadataObj.XbDetail.RoutingValue,
		StatusHistories:     buildStatusHistories(disbursement.StatusHistories),
	}
	return &response
}

func buildStatusHistories(histories []*statusHistoriesModel.StatusHistory) []xbModel.XbPayoutStatusHistoryResponse {
	if len(histories) <= 1 {
		return []xbModel.XbPayoutStatusHistoryResponse{}
	}
	sort.Slice(histories, func(i, j int) bool {
		return histories[i].CreatedAt.Before(histories[j].CreatedAt)
	})

	histories = histories[1:]

	length := len(histories)
	labels := map[string]int{}
	result := make([]xbModel.XbPayoutStatusHistoryResponse, 0, length)

	for i, history := range histories {
		next := i + 1
		nextStatusIsInfoRequested := length > next && histories[next].Status == constant.XbStatusInfoRequested

		// got that info_requested status was overalapping with processing, even though its already sorted
		// should remove the processing status before info_requested to avoid duplication
		if history.Status == constant.XbStatusExpired || nextStatusIsInfoRequested {
			continue
		}

		statusHistory := xbModel.XbPayoutStatusHistoryResponse{
			Label:          history.MetadataObj.Label,
			Description:    history.MetadataObj.Description,
			Recommendation: history.MetadataObj.Recommendation,
			Timestamp:      history.CreatedAt,
		}

		if idx, ok := labels[statusHistory.Label]; ok {
			result[idx] = statusHistory

		} else {
			result = append(result, statusHistory)
			labels[statusHistory.Label] = len(result) - 1
		}
	}
	return result
}
