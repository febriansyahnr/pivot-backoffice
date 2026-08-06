package payoutManualProcessingAccount

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	payoutManualProcessingAccountModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payoutManualProcessingAccount"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/google/uuid"
)

func (c *CRMPayoutManualProcessingAccountController) List(w http.ResponseWriter, r *http.Request) {
	ctx, segment := otelTracer.Start(r.Context(), "port/http/controller/v1/crmController/payoutManualProcessingAccount/List")
	defer segment.End()

	var (
		page     = constant.DefaultPage
		pageSize = constant.DefaultPaginationPageSize
		err      error
	)

	query := r.URL.Query()
	var payload payoutManualProcessingAccountModel.PayoutManualProcessingAccountQuery

	if merchantIDStr := query.Get("merchantId"); merchantIDStr != "" {
		merchantID, parseErr := uuid.Parse(merchantIDStr)
		if parseErr != nil {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
			return
		}
		payload.MerchantID = merchantID
	}

	payload.BankCode = query.Get("bankCode")
	payload.AccountNumber = query.Get("accountNumber")
	payload.Status = query.Get("status")
	if payload.Status != "" && payload.Status != constant.StatusActive && payload.Status != constant.StatusInactive {
		response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPayoutManualProcessingAccountStatus))
		return
	}

	payload.SortBy = query.Get("sortBy")
	payload.Sort = query.Get("sort")

	strPage := query.Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := query.Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendOpenApiResponseError(w, errors.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	result, err := c.service.List(ctx, &payload)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	list := []*payoutManualProcessingAccountModel.PayoutManualProcessingAccountResponse{}
	for _, account := range result.Data.([]*payoutManualProcessingAccountModel.PayoutManualProcessingAccount) {
		list = append(list, account.ToResponse())
	}
	result.Data = list

	response.SendOpenApiResponseOK(w, result)
}
