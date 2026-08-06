package middleware

import (
	"github.com/paper-indonesia/pivot-backoffice/constant"
	inboundPdk "github.com/paper-indonesia/pdk/v2/chiExt/inbound"
	pdkMiddleware "github.com/paper-indonesia/pdk/v2/chiExt/middleware"
	"net/http"
)

func InboundFeatureMiddleware(feature string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			inboundClient := &inboundPdk.HttpClientDetails{
				Feature: feature,
			}

			if merchantId, ok := r.Context().Value(constant.CtxMerchantIDKey).(string); ok {
				inboundClient.ReferenceId = merchantId
			}
			if inboundClient.ReferenceId == "" {
				inboundClient.ReferenceId = r.Header.Get(constant.ClientIdKey)
			}

			if rw, ok := w.(*pdkMiddleware.ResponseWriter); ok {
				rw.SetClientDetails(inboundClient)
			}

			next.ServeHTTP(w, r)
		})
	}
}
