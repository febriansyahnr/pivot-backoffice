package xbPayoutController

import (
	errs "errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	disbursementModel "github.com/paper-indonesia/pivot-backoffice/internal/model/disbursement"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	xbModel "github.com/paper-indonesia/pivot-backoffice/internal/model/xb"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/go/errors"
	"github.com/shopspring/decimal"
)

func (c *xbPayoutController) GetList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/xbPayout/GetList")
	defer segment.End()

	var (
		page    int64 = 1
		perPage int64 = constant.DefaultPaginationPageSize
	)

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	// Get query params
	status := r.URL.Query().Get("status")
	uuid := r.URL.Query().Get("uuid")
	sortBy := r.URL.Query().Get("sortBy")
	sort := r.URL.Query().Get("sort")

	// Validation and parsing
	if err := c.bindOptionalInt64Query("page", r, &page); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err := c.bindOptionalInt64Query("perPage", r, &perPage); err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	startAt, endAt, err := c.getQueryForXbPayoutList(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	filter := &disbursementModel.GetDisbursementFilterRequest{
		MerchantID:     user.MerchantId,
		UUID:           uuid,
		StartCreatedAt: &startAt,
		EndCreatedAt:   &endAt,
		ReasonType:     status, // XB status = disbursement.reason_type
		SortBy:         sortBy,
		Sort:           sort,
		IsXbPayout:     true,
	}

	list, err := c.disbursementSvc.GetList(ctx, filter, page, perPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	resp, err := buildDataResponse(list.Data)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, resp, list.Meta)
}

func (c *xbPayoutController) bindOptionalInt64Query(key string, r *http.Request, dst *int64) (err error) {
	val := r.URL.Query().Get(key)

	if val == "" {
		return

	} else if dst == nil {
		return errs.New("dst value can't be nil")
	}

	if *dst, err = strconv.ParseInt(val, 10, 64); err != nil {
		return pkgErrors.New(response.HttpErrRequest, fmt.Errorf("invalid %s format. Use number format instead", key))
	}
	return nil
}

func (c *xbPayoutController) getQueryForXbPayoutList(r *http.Request) (startAt, endAt time.Time, err error) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	if startDate == "" || endDate == "" {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, errors.New("start date and end date must be filled"))
	}

	if startAt, err = time.ParseInLocation(constant.DateFormat, startDate, loc); err != nil {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidStartDateFmt)
	}

	if endAt, err = time.ParseInLocation(constant.DateFormat, endDate, loc); err != nil {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrInvalidEndDateFmt)
	}

	if startAt.After(endAt) {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, constant.ErrFilterDateInput)

	} else if (startAt.Sub(endAt).Hours() / 24) > 31 {
		return startAt, endAt, pkgErrors.New(response.HttpErrRequest, errors.New("maximum date range 31 days"))
	}

	return startAt.UTC(), endAt.Add(24 * time.Hour).UTC(), nil
}

var loc, _ = time.LoadLocation(constant.TimeLoc)

func buildDataResponse(listData interface{}) (resp []*xbModel.GetXbPayoutListResponse, err error) {
	disbursementDataList, ok := listData.([]*disbursementModel.DisbursementWithTransactionResponse)
	if !ok {
		return nil, pkgErrors.New(response.HttpErrInternal, errors.New("failed to map response"))
	}

	tempResp := make([]*xbModel.GetXbPayoutListResponse, len(disbursementDataList))
	for idx, disbursement := range disbursementDataList {
		fee := decimal.NewFromFloat(0)
		if disbursement.Fee != nil {
			fee = *disbursement.Fee
		}

		sourceAmount := disbursement.MetadataObj.XbDetail.SourceAmount
		totalAmount := disbursement.MetadataObj.XbDetail.TotalAmount

		status := ""
		if disbursement.ReasonType != nil {
			status = *disbursement.ReasonType
		}

		tempResp[idx] = &xbModel.GetXbPayoutListResponse{
			UUID:                   disbursement.UUID,
			ReferenceID:            disbursement.ReferenceID,
			SourceCurrency:         disbursement.MetadataObj.XbDetail.SourceCurrency,
			DestinationCurrency:    disbursement.Currency,
			DestinationAmount:      disbursement.Amount,
			SourceAmount:           sourceAmount,
			Fee:                    fee,
			TotalAmount:            totalAmount,
			CreatedAt:              disbursement.CreatedAt,
			Status:                 status,
			BeneficiaryAccountName: disbursement.BeneficiaryAccountName,
		}
	}

	return tempResp, nil
}
