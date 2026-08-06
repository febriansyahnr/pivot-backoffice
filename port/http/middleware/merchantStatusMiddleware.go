package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/service"
	services "github.com/paper-indonesia/pivot-backoffice/internal/service"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

type ErrResponseWrapperFunc func(ctx context.Context, w http.ResponseWriter, err error)

func MerchantStatusMiddleware(merchantSvc services.IMerchantService, cfg *config.Config, errResponseWrapper ErrResponseWrapperFunc) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				ctx, segment := otelTracer.Start(r.Context(), "port/http/middleware/MerchantStatusMiddleware")
				defer segment.End()

				merchantId, _ := ctx.Value(constant.CtxMerchantIDKey).(string)
				if merchantId == "" {
					return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrMerchantIDNotValid)
				}

				if err := validateMerchantStatus(ctx, r, cfg, merchantSvc, merchantId); err != nil {
					return err
				}

				subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID)
				if subMerchantId == "" {
					return nil
				}
				return validateMerchantStatus(ctx, r, cfg, merchantSvc, subMerchantId)
			}()
			if err != nil {
				ctx := r.Context()
				if subMerchantId := r.Header.Get(constant.HeaderXSubMerchantID); subMerchantId != "" {
					ctx = context.WithValue(ctx, constant.CtxSubMerchantIDKey, subMerchantId)
				}
				errResponseWrapper(ctx, w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func validateMerchantStatus(ctx context.Context, r *http.Request, cfg *config.Config, merchantSvc service.IMerchantService, merchantId string) error {
	merchantStatus, err := merchantSvc.GetCachedMerchantStatus(ctx, merchantId)
	if err != nil {
		return pkgErrs.New(response.HttpErrUnauthorized, constant.ErrMerchantNotFound)
	}

	switch merchantStatus.Status {
	case constant.MerchantStatusDormant:
		if !cfg.MerchantConfig.DormantValidation || r.Method == http.MethodGet {
			return nil
		}
		return pkgErrs.New(response.HttpErrForbidden,
			fmt.Errorf("perform is not permitted caused by %s", merchantStatus.ReasonStatus))

	case constant.MerchantStatusClosed:
		return pkgErrs.New(response.HttpErrUnauthorized,
			fmt.Errorf("Merchant status is closed. Reason: %s", merchantStatus.ReasonStatus))

	case constant.MerchantStatusCreated:
		// Will be removed when we have merchant onboarding process feature
		return pkgErrs.New(response.HttpErrForbidden,
			errors.New("Merchant status is not active yet, please contact the operations team to be able to use this feature."))

	case constant.MerchantStatusBlocked:
		return pkgErrs.New(response.HttpErrForbidden,
			errors.New("Merchant status is blocked, please contact the operations team to be able to use this feature."))

	case constant.MerchantStatusDeactivated:
		return pkgErrs.New(response.HttpErrUnauthorized,
			errors.New("Merchant status is deactivated by parent merchant, please contact your parent merchant to activate the merchant status."))
	}
	return nil
}
