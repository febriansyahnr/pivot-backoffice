package transfer

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	transferModel "github.com/paper-indonesia/pivot-backoffice/internal/model/transfer"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// GetList		godoc
// @Summary		provide list of payment transfer histories of the merchant
// @Description	this endpoint will contain information about transfer histories of the merchant, including the amount, recipient, and status of the transfer
// @ID			payment-transfer-list
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Success		200  	{object}				response.ApiResponse{data=[]transferModel.Transfer, meta=commonModel.Meta}
// @Failure		500  	{object}				response.ApiResponse
// @Router		/api/v1/transfers 	[get]
// @Security 	Bearer
func (c *TransferController) FilterTransferHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/transfer/FilterTransferHistory")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	request, err := c.ParseFilterParam(r)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, err))
		return
	}

	request.MerchantID = user.MerchantId
	result, err := c.transferService.GetList(ctx, &request, request.Page, request.PerPage)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponsePaginationOK(w, result.Data, result.Meta)
}

// PaymentHistory		godoc
// @Summary		Payment transfer detail Endpoint
// @Description	Get Detail Of Payment Transfer Activity, it will contain the amount, recipient, and status of the transfer
// @ID			api-payment-transfer-detail
// @Tags		API - Payment
// @Accept		json
// @Produce		json
// @Param 		id		path		string true "Transfer ID"
// @Success		200  	{object}	response.ApiResponse{data=transferModel.transfer.TransferTransactionDetail}
// @Failure		500  	{object}	response.ApiResponse
// @Router		/api/v1/transfers/{id} [get]
// @Security	Bearer
func (c *TransferController) GetTransferByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v1/transfer/GetTransferByID")
	defer segment.End()

	// Get User Info from jwt token
	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrUnauthorized, constant.ErrUserNotFound))
		return
	}

	transferID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(transferID); err != nil {
		response.SendApiResponseError(ctx, w, pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid transfer ID format")))
		return
	}

	req := transferModel.GetTransferTransactionRequest{
		MerchantID:    user.MerchantId,
		TransactionID: transferID,
	}

	if r.URL.Query().Get("subMerchantId") != "" {
		err := c.merchantService.ValidateSubMerchantParent(ctx, user.MerchantId, r.URL.Query().Get("subMerchantId"))
		if err != nil {
			response.SendApiResponseError(ctx, w, err)
			return
		}

		req.MerchantID = r.URL.Query().Get("subMerchantId")
		req.ParentID = user.MerchantId
	}

	transfer, err := c.transferService.GetTransferTransaction(ctx, req)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, transfer)
}

// ParseFilterParam parses the filter parameters from the HTTP request and returns a GetTransferListRequest object.
// It sets default values for SortOrder, SortBy, Page, and PerPage if they are not provided in the request.
// The function validates and parses the following query parameters:
// - page: the page number (default is 1).
// - perPage: the number of items per page (default is 10).
// - startDate: the start date in 'YYYY-MM-DDTHH:mm:ssZ' format.
// - endDate: the end date in 'YYYY-MM-DDTHH:mm:ssZ' format.
// - sort: the sort order (default is "ASC").
// - sortBy: the field to sort by (default is "createdAt").
// - status: the status filter.
// - uuid: the UUID filter.
// If any parameter is invalid, it returns an error with a descriptive message.
func (c *TransferController) ParseFilterParam(r *http.Request) (transferModel.GetTransferListRequest, error) {
	var (
		opt transferModel.GetTransferListRequest
	)
	opt.SortOrder = "ASC"
	opt.SortBy = "createdAt"
	opt.Page = 1
	opt.PerPage = 10

	if r.URL.Query().Get("page") != "" {
		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid page format. Use number format instead"))
		}
		opt.Page = int64(page)
	}

	if r.URL.Query().Get("perPage") != "" {
		perPage, err := strconv.Atoi(r.URL.Query().Get("perPage"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid perPage format. Use number format instead"))
		}
		opt.PerPage = int64(perPage)
	}

	if r.URL.Query().Get("startDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("startDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid startDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.StartDate = d
	}

	if r.URL.Query().Get("endDate") != "" {
		d, err := time.Parse(util.UTCLayout, r.URL.Query().Get("endDate"))
		if err != nil {
			return opt, pkgErr.New(response.HttpErrRequest, fmt.Errorf("invalid endDate format. Use 'YYYY-MM-DDTHH:mm:ssZ' format"))
		}
		opt.EndDate = d
	}

	if r.URL.Query().Get("sort") != "" && allowedSortOrders[r.URL.Query().Get("sort")] {
		opt.SortOrder = r.URL.Query().Get("sort")
	}

	if r.URL.Query().Get("sortBy") != "" && allowedSort[r.URL.Query().Get("sortBy")] {
		opt.SortBy = r.URL.Query().Get("sortBy")
	}

	opt.Status = r.URL.Query().Get("status")
	opt.UUID = r.URL.Query().Get("uuid")
	opt.PaymentReferenceID = opt.UUID
	opt.PaymentID = opt.UUID

	return opt, nil
}
