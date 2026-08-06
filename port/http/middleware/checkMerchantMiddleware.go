package middleware

import (
	"context"
	"net/http"

	merchantModel "github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"

	"github.com/paper-indonesia/pivot-backoffice/internal/service"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	pdkConstant "github.com/paper-indonesia/pdk/v2/constant"
)

func CheckMerchantMiddleware(merchantSvc service.IMerchantService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctxReq := r.Context()

			err := func() error {
				ctx, segment := otelTracer.Start(ctxReq, "http/middleware/CheckMerchantMiddleware")
				defer segment.End()

				merchantID, _ := ctx.Value(constant.CtxMerchantIDKey).(string)
				if merchantID == "" {
					merchantID = r.Header.Get(constant.HeaderXMerchantId)
				}

				if _, err := uuid.Parse(merchantID); err != nil {
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrIncorrectMerchantID)
				}

				merchant, err := merchantSvc.FindMerchantByID(ctx, merchantID)
				if err != nil {
					return pkgErrors.New(response.HttpErrInternal, constant.ErrFailedValidation)
				} else if merchant == nil {
					return pkgErrors.New(response.HttpErrUnauthorized, constant.ErrIncorrectMerchantID)
				}

				// Set Derived MerchantID
				ctxReq = setDerivedMerchantID(ctxReq, merchant)

				// Add merchant id to context
				ctxReq = context.WithValue(ctxReq, constant.CtxMerchantIDKey, merchantID)

				return nil
			}()

			if err != nil {
				response.SendOpenApiResponseError(w, err)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctxReq))
		})
	}
}

func setDerivedMerchantID(ctx context.Context, merchant *merchantModel.Merchant) context.Context {
	if merchant.ParentID.Valid && merchant.KYCStatus.String == constant.KYCStatusNotRequired {
		ctx = context.WithValue(ctx, constant.CtxDerivedMerchantID, merchant.ParentID.String)
		ctx = context.WithValue(ctx, constant.CtxParentMerchantId, merchant.ParentID.String)
		ctx = context.WithValue(ctx, pdkConstant.CtxSubMerchantIDKey, merchant.UUID)
	}

	return ctx
}
