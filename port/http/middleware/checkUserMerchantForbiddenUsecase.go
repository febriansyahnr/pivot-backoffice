package middleware

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func CheckUserMerchantForbiddenUsecase(forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService, usecase string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/CheckUserMerchantForbiddenUsecase")
				defer segment.End()

				if r.Method == http.MethodGet {
					return nil
				}

				// Get User Info from jwt token
				user, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
				if !ok {
					return errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
				}

				if err := forbiddenUsecaseSvc.CheckUseCase(ctx, user.MerchantId, usecase); err != nil {
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
