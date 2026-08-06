package ledgerController

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	commonModel "github.com/paper-indonesia/pivot-backoffice/internal/model/common"
	ledger_model "github.com/paper-indonesia/pivot-backoffice/internal/model/ledger"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *LedgerController) GetTransactions(w http.ResponseWriter, r *http.Request) {
	var (
		ctx        = r.Context()
		req        ledger_model.GetLedgerTransactionRequest
		pagination commonModel.Meta
		err        error
	)
	timeLoc, _ := time.LoadLocation(constant.TimeLoc)

	ctx, segment := otelTracer.Start(ctx, "port/http/controller/v2/ledger/GetTransactions")
	defer segment.End()

	accountID := r.URL.Query().Get("accountId")
	if errId := uuid.Validate(accountID); errId != nil {
		response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, errors.New("invalid accountId")))
		return
	}
	req.AccountID, _ = uuid.Parse(accountID)

	referenceType := r.URL.Query().Get("referenceType")
	req.ReferenceType = referenceType

	startDate := r.URL.Query().Get("startDate")
	if startDate != "" {
		date, err := time.ParseInLocation(constant.DateFormat, startDate, timeLoc)
		if err != nil {
			response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		req.StartDate = date
	}

	endDate := r.URL.Query().Get("endDate")
	if endDate != "" {
		date, err := time.ParseInLocation(constant.DateFormat, endDate, timeLoc)
		if err != nil {
			response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, err))
			return
		}
		req.EndDate = date
	}

	currentPage := r.URL.Query().Get("page")
	if currentPage == "" {
		pagination.Page = constant.DefaultPage
	}
	if currentPage != "" {
		page, err := strconv.Atoi(currentPage)
		if err != nil {
			response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, errors.New("failed parse page to integer")))
			return
		}
		if page < constant.DefaultPage {
			page = constant.DefaultPage
		}
		pagination.Page = int64(page)
	}
	pageSize := r.URL.Query().Get("pageSize")
	if pageSize == "" {
		pagination.PerPage = int64(constant.DefaultPaginationPageSize)
	}
	if pageSize != "" {
		size, err := strconv.Atoi(pageSize)
		if err != nil {
			response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrRequest, errors.New("failed parse pageSize to integer")))
			return
		}
		if size < constant.DefaultPaginationPageSize {
			size = constant.DefaultPaginationPageSize
		}
		pagination.PerPage = int64(size)
	}

	responseData, err := c.ledgerSvc.GetLedgerTransactions(ctx, &req, &pagination)
	if err != nil {
		response.SendOpenApiResponseError(w, err)
		return
	}

	responseData.Data = ToGetTransactionsResponse(responseData.Data.([]*ledger_model.GetLedgerTransactionData))
	response.SendApiResponseOK(w, responseData)
}

type GetTransactionsResponse struct {
	ReferenceID          string    `json:"referenceId"`
	Debit                float64   `json:"debit"`
	Credit               float64   `json:"credit"`
	Type                 string    `json:"type"`
	Channel              string    `json:"channel"`
	Status               string    `json:"status"`
	Remarks              string    `json:"reason"`
	ReasonType           string    `json:"reasonType"`
	ReasonDescription    string    `json:"reasonDescription"`
	TransactionTimestamp time.Time `json:"transactionTimestamp"`
}

func ToGetTransactionsResponse(data []*ledger_model.GetLedgerTransactionData) []*GetTransactionsResponse {
	var res []*GetTransactionsResponse
	for _, d := range data {
		res = append(res, &GetTransactionsResponse{
			ReferenceID:          d.ReferenceID,
			Debit:                d.Debit,
			Credit:               d.Credit,
			Type:                 d.Type,
			Channel:              d.Channel,
			Status:               d.Status,
			Remarks:              d.Remarks,
			ReasonType:           d.ReasonType,
			ReasonDescription:    d.ReasonDescription,
			TransactionTimestamp: d.TransactionTimestamp,
		})
	}
	return res
}
