package httputil

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func BindMerchantID(r *http.Request, merchantId *string) {
	merchId := r.Header.Get(constant.HeaderXMerchantId)
	if merchId != "" {
		*merchantId = merchId
	}
}
