package middleware

import (
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	userModel "github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func MerchantUserProductValidationMiddleware(productSvc service.IProductService, productName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "http/middleware/MerchantUserProductValidationMiddleware")
				defer segment.End()

				userClaims, ok := ctx.Value(constant.CtxUserInfoKey).(*userModel.UserTokenClaims)
				if !ok {
					return errors.New(response.HttpErrUnauthorized, constant.ErrUserNotFound)
				}

				if err := productSvc.ValidateMerchantProductAvailability(ctx, &product.ValidateMerchantProductAvailability{
					MerchantID:  userClaims.MerchantId,
					ProductName: productName,
				}); err != nil {
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

func MerchantProductValidationMiddleware(productSvc service.IProductService, productName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, segment := otelTracer.Start(r.Context(), "http/middleware/MerchantProductValidationMiddleware")
			defer segment.End()

			// Get claims from merchant auth
			merchantClaims, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
			if !ok || merchantClaims == nil {
				response.SendOpenApiResponseError(w, pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken))
				return
			}

			err := productSvc.ValidateMerchantProductAvailability(ctx, &product.ValidateMerchantProductAvailability{
				MerchantID:  merchantClaims.MerchantId,
				ProductName: productName,
			})
			if err != nil {
				response.SendOpenApiResponseError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
