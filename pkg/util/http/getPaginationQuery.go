package httputil

import (
	"net/http"
	"strconv"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func GetPaginationQuery(r *http.Request) (int64, int64) {
	page := constant.DefaultPage
	perPage := constant.DefaultPaginationPageSize
	if p, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(r.URL.Query().Get("perPage")); err == nil && pp > 0 {
		perPage = pp
	}
	return int64(page), int64(perPage)
}
