package merchant

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	tncModel "github.com/paper-indonesia/pivot-backoffice/internal/model/tnc"
	pkgErr "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func (c *CRMMerchantController) GetMerchantTNCHistory(w http.ResponseWriter, r *http.Request) {
	merchantID := r.PathValue("id")
	if merchantID == "" {
		response.SendApiResponseError(r.Context(), w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidMerchantID))
		return
	}

	var (
		page     = constant.DefaultPage
		pageSize = constant.DefaultPaginationPageSize
		err      error
	)

	query := r.URL.Query()
	payload := tncModel.SigningHistoryQuery{
		MerchantID: merchantID,
		TNCVersion: query.Get("version"),
	}

	strPage := query.Get("page")
	if strPage != "" {
		page, err = strconv.Atoi(strPage)
		if err != nil || page < 1 {
			response.SendApiResponseError(r.Context(), w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidPage))
			return
		}
	}
	payload.Page = int64(page)

	strPageSize := query.Get("perPage")
	if strPageSize != "" {
		pageSize, err = strconv.Atoi(strPageSize)
		if err != nil || pageSize < 1 {
			response.SendApiResponseError(r.Context(), w, pkgErr.New(response.HttpErrRequest, constant.ErrInvalidPerPage))
			return
		}
	}
	payload.PageSize = int64(pageSize)

	result, err := c.tncSvc.GetSigningHistory(r.Context(), &payload)
	if err != nil {
		response.SendApiResponseError(r.Context(), w, err)
		return
	}

	response.SendApiResponseOK(w, result)
}
