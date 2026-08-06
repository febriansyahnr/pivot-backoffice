package middleware

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
)

func PlatformPermissionOption(nonPlatformPermissions MiddlewareFunc, permissionSvc service.IPermissionService, permissions ...string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Platform Permissions
			if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
				PermittedPermissions(permissionSvc, permissions...)(next).ServeHTTP(w, r)
				return
			}

			// Non Platform Permissions
			nonPlatformPermissions(next).ServeHTTP(w, r)
		})
	}
}
