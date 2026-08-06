package httputil

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
)

func BindSubmerchantID(r *http.Request, merchantId *string) {
	submerchId := r.Header.Get(constant.HeaderXSubMerchantID)
	if submerchId != "" {
		*merchantId = submerchId
	}
}

func BindLoggedInUserType(r *http.Request, userType *string) {
	submerchId := r.Header.Get(constant.HeaderXSubMerchantID)
	*userType = constant.UserTypeMerchant
	if submerchId != "" {
		*userType = constant.UserTypeSubMerchant
	}
}
