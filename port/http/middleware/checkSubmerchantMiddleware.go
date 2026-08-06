package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/product"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func CheckSubMerchantMiddleware(merchantSvc service.IMerchantService, productSvc service.IProductService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := r.Context()

			err := func() error {
				ctx, segment := otelTracer.Start(ctxReq, "http/middleware/CheckSubMerchantMiddleware")
				defer segment.End()

				merchant, ok := ctx.Value(constant.CtxMerchantInfo).(*merchant.MerchantAuthTokenClaims)
				if !ok || merchant == nil {
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrInvalidToken)
				}

				submerchId := r.Header.Get(constant.HeaderXSubMerchantID)
				if submerchId == "" {
					return nil
				}

				// Append parent merchantID
				ctxReq = context.WithValue(ctxReq, constant.CtxParentMerchantId, merchant.MerchantId)

				if _, err := uuid.Parse(submerchId); err != nil {
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrIncorrectMerchantID)
				}
				subMerchant, err := merchantSvc.FindMerchantByID(ctx, submerchId)
				if err != nil {
					return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidation)
				}
				if subMerchant == nil {
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrIncorrectMerchantID)
				}

				if subMerchant.ParentID.String != merchant.MerchantId {
					return pkgErrors.New(response.HttpErrForbidden, constant.ErrForbiddenAccess)
				}

				// Set Derived MerchantID
				ctxReq = setDerivedMerchantID(ctxReq, subMerchant)

				if errValidate := productSvc.ValidateMerchantProductAvailability(ctx, &product.ValidateMerchantProductAvailability{
					MerchantID:  merchant.MerchantId,
					ProductName: constant.ProductPlatform,
				}); errValidate != nil {
					return errValidate
				}

				return nil
			}()

			if err != nil {
				ctxReq = context.WithValue(ctxReq, constant.CtxCustomErrorResponse, response.OpenApiErrorResponseType1(response.SendOpenApiResponseError))
				if errors.Is(err, constant.ErrForbiddenAccess) {
					ctxReq = context.WithValue(ctxReq, constant.CtxErrorInfo, pkgErrors.New(response.HttpErrForbidden, errors.New(constant.ErrMessageV2RequestForbidden)))
				}
				response.SendOpenApiNonSnapResponseError(ctxReq, w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctxReq))
		})
	}
}
