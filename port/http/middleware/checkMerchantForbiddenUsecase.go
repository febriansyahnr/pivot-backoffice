package middleware

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httputil "github.com/paper-indonesia/pivot-backoffice/pkg/util/http"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func CheckMerchantForbiddenUsecase(forbiddenUsecaseSvc service.IMerchantForbiddenUseCaseService, usecase string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/CheckMerchantForbiddenUsecase")
				defer segment.End()

				if r.Method == http.MethodGet {
					return nil
				}

				// Get User Info from jwt token
				merchantInfoFromCtx := ctx.Value(constant.CtxMerchantInfo)
				merchant, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
				if !ok {
					return errors.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound)
				}

				merchantIdToCheck := merchant.MerchantId
				httputil.BindSubmerchantID(r, &merchantIdToCheck)

				if err := forbiddenUsecaseSvc.CheckUseCase(ctx, merchantIdToCheck, usecase); err != nil {
					return err
				}

				return nil
			}()

			if err != nil {
				response.SendOpenApiResponseError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
