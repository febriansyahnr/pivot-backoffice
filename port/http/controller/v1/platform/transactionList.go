package platform

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/platform"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	errPkg "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *PlatformController) TransactionList(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "/port/http/controller/v1/platform/TransactionList")
	defer segment.End()

	user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
	if !ok {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrUnauthorized, constant.ErrInvalidAccess))
		return
	}
	request := platform.TransactionRequest{
		ParentMerchantId: user.MerchantId,
	}

	if err := httputil.ValidateReportDateRangeFromRequest(r, "startDate", "endDate"); err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, err))
		return
	}

	submerchantId := r.URL.Query().Get("merchantId")
	_, err := uuid.Parse(submerchantId)
	if err != nil {
		response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, constant.ErrInvalidMerchantId))
		return
	}
	request.MerchantId = submerchantId
	request.Reference = r.URL.Query().Get("reference")
	request.ReferenceType = r.URL.Query().Get("referenceType")
	request.ReferenceID = r.URL.Query().Get("referenceId")
	request.Status = r.URL.Query().Get("status")
	request.ApprovalStatus = r.URL.Query().Get("approvalStatus")
	request.PaymentMethod = r.URL.Query().Get("paymentMethod")
	request.Keyword = r.URL.Query().Get("keyword")
	request.UUID = r.URL.Query().Get("uuid")

	// start & enddatetime
	startDate := r.URL.Query().Get("startDate")
	if startDate != "" {
		startTime, err := time.Parse(util.UTCLayout, startDate)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid startDate value")))
			return
		}
		request.StartDate = startTime
	}

	endDate := r.URL.Query().Get("endDate")
	if endDate != "" {
		endTime, err := time.Parse(util.UTCLayout, endDate)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid endDate value")))
			return
		}
		request.EndDate = endTime
	}

	paymentStartDate := r.URL.Query().Get("paymentStartDate")
	if paymentStartDate != "" {
		startTime, err := time.Parse(util.UTCLayout, paymentStartDate)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid paymentStartDate value")))
			return
		}
		request.PaymentStartDate = startTime
	}

	paymentEndDate := r.URL.Query().Get("paymentEndDate")
	if paymentEndDate != "" {
		endTime, err := time.Parse(util.UTCLayout, paymentEndDate)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid paymentEndDate value")))
			return
		}
		request.PaymentEndDate = endTime
	}

	// sort
	sortCol := r.URL.Query().Get("sortBy")
	if sortCol != "" {
		request.SortBy = sortCol
	}
	request.SortOrder = constant.SortOrderAsc
	sortOrder := r.URL.Query().Get("sort")
	if sortOrder != "" {
		request.SortOrder = sortOrder
	}

	// page
	request.Page = constant.DefaultPage
	page := r.URL.Query().Get("page")
	if page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid page value")))
			return
		}
		request.Page = int64(p)
	}

	request.PerPage = constant.DefaultPageSize
	perPage := r.URL.Query().Get("perPage")
	if perPage != "" {
		p, err := strconv.Atoi(perPage)
		if err != nil {
			response.SendApiResponseError(ctx, w, errPkg.New(response.HttpErrRequest, errors.New("invalid perPage value")))
			return
		}
		request.PerPage = int64(p)
	}

	responseData, err := c.platformSvc.GetMerchantTransactionList(ctx, &request)
	if err != nil {
		response.SendApiResponseError(ctx, w, err)
		return
	}

	response.SendApiResponseOK(w, responseData)
}
