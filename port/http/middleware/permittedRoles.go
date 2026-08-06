package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func PermittedRoles(roles ...string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/PermittedRoles")
				defer segment.End()

				user, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
				if !ok {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
				}

				ok = slices.ContainsFunc(
					roles, func(role string) bool {
						return strings.EqualFold(role, user.Role)
					},
				)
				if !ok {
					return pkgErrs.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
				}

				return nil
			}()

			if err != nil {
				response.SendApiResponseError(r.Context(), w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
