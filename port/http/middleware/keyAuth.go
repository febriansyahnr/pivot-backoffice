package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func KeyAuth(key string, values ...string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				_, segment := otelTracer.Start(r.Context(), "http/middleware/KeyAuth")
				defer segment.End()

				auth := r.Header.Get(key)
				if auth = strings.TrimSpace(auth); auth == "" {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrKeyAuthRequired)
				}

				if !slices.Contains(values, auth) {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrInvalidKey)
				}

				return nil
			}()

			if err != nil {
				response.SendGeneralResponseError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
