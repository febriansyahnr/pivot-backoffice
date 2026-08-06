package middleware

import (
	"context"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"net/http"
)

func SetV2ErrorCodeMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), constant.CtxUseV2ErrorCode, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetErrorSourceMiddleware adds error.source (UPSTREAM/DOWNSTREAM/SYSTEM) to error responses
// without applying V2 error code remapping. Use this for V1 endpoints that need error source
// classification but should keep their original response codes.
func SetErrorSourceMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), constant.CtxUseErrorSource, true)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
