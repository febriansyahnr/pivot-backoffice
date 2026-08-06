package middleware

import (
	"net/http"
	"slices"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func EnvironmentCheck(env string, allowedEnv ...string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				_, segment := otelTracer.Start(r.Context(), "http/middleware/EnvironmentCheck")
				defer segment.End()

				if !slices.Contains(allowedEnv, env) {
					return pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
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
