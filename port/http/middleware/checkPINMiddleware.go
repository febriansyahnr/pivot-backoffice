package middleware

import (
	"encoding/base64"
	e "errors"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/pkg/rabbitMqExt"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func CheckPINMiddleware(
	userSvc service.IUserService,
	rabbitMq rabbitMqExt.IRabbitMQExt) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/CheckPINMiddleware")
				defer segment.End()

				claims, ok := ctx.Value(constant.CtxUserInfoKey).(*user.UserTokenClaims)
				if !ok {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
				}

				reqEncodedPin := r.Header.Get(constant.HeaderXRequestPIN)
				if reqEncodedPin == "" {
					return pkgErrs.New(response.HttpErrRequest, constant.ErrRequiredPIN)
				}

				pin, _ := base64.StdEncoding.DecodeString(reqEncodedPin)
				if err := userSvc.CheckCurrentPin(ctx, claims.UUID, string(pin)); err != nil {
					if e.Is(err, constant.ErrRateLimiterExceedFailedAttempts) {
						// publish activity, do nothing on error
						_ = rabbitMq.PublishActivity(
							ctx,
							&claims.MerchantId,
							&claims.UUID,
							constant.TagAccount,
							constant.ErrFailedCheckPINLimit.Error(),
							map[string]string{
								"email": claims.Email,
							},
						)
					}

					return err
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
