package openApi

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/merchant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/logger"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/paper-indonesia/pdk/v2/idempotenshine"
	ffclient "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

type IdempotentMiddleware struct {
	logger logger.ILogger
	cfg    config.Config
	sc     config.Secret
	redis  redisExt.IRedisExt
}

func NewIdempotentMiddleware(logger logger.ILogger,
	cfg config.Config,
	sc config.Secret,
	redis redisExt.IRedisExt) *IdempotentMiddleware {
	return &IdempotentMiddleware{logger, cfg, sc, redis}
}

func (m *IdempotentMiddleware) SnapIdempotentMiddleware(pathName string) func(http.Handler) http.Handler {
	return idempotenshine.Middleware(
		m.cfg.ServiceName,
		pathName,
		idempotenshine.WithTTL(24*time.Hour),
		idempotenshine.WithKeySource(idempotenshine.HeaderKeySource("X-EXTERNAL-ID")),
		idempotenshine.WithRedisClient(m.redis.Client()),
		idempotenshine.WithLogicOptionUsage(idempotenshine.LogicOption{
			ReturnErrorWhenKeyExists: true,
		}),
		idempotenshine.WithConflictResponder(snapErrResponse()),
		idempotenshine.WithClientResponder(snapErrResponse()),
		idempotenshine.WithServerResponder(snapErrResponse()),
	)
}

func (m *IdempotentMiddleware) internalIdempotencyProcess(merchantInfo *merchant.MerchantAuthTokenClaims, pathName string, keySource string) func(http.Handler) http.Handler {
	headerKey := idempotenshine.HeaderKeySource(strings.ToUpper(keySource))

	option := []idempotenshine.Option{
		idempotenshine.WithTTL(24 * time.Hour),
		idempotenshine.WithKeySource(headerKey),
		idempotenshine.WithRedisClient(m.redis.Client()),
		idempotenshine.WithLogicOptionUsage(idempotenshine.LogicOption{
			ReturnDataWhenKeyExistsWithBodyValidation: true,
		}),
		idempotenshine.WithConflictResponder(conflictErrorHandler(keySource)),
		idempotenshine.WithServerResponder(serverErrorHandler(keySource)),
	}

	if isMerchantExcludeKeyValidation(merchantInfo) {
		// use this ff until all merchant have move their idempotency key value
		option = append(option,
			idempotenshine.WithClientResponder(clientErrorHandler(keySource, false)),
		)
	} else {
		option = append(option,
			idempotenshine.WithValidationKey(idempotenshine.SpecialAlphanumeric, 16, 36),
			idempotenshine.WithClientResponder(clientErrorHandler(keySource, true)),
		)
	}

	return idempotenshine.Middleware(
		m.cfg.ServiceName,
		pathName,
		option...,
	)
}

func (m *IdempotentMiddleware) InternalIdempotencyMiddleware(pathName string, keySource string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		// get context from request
		var (
			merchantInfo *merchant.MerchantAuthTokenClaims
		)
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Get User Info from jwt token
			merchantInfoFromCtx := req.Context().Value(constant.CtxMerchantInfo)
			data, ok := merchantInfoFromCtx.(*merchant.MerchantAuthTokenClaims)
			if !ok {
				return
			}

			merchantInfo = data
			handler := m.internalIdempotencyProcess(merchantInfo, pathName, keySource)(next)
			handler.ServeHTTP(w, req)
		})

	}
}

func snapErrResponse() idempotenshine.ResponseHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		response.SendOpenApiSnapResponseError(ctx, w, pkgErrs.New(response.SnapErrConflict, fmt.Errorf("%s", constant.ConflictErrMsg)))
	}
}

func conflictErrorHandler(headerKey string) idempotenshine.ResponseHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := r.Header.Get(headerKey)
		if key == "" {
			response.SendOpenApiNonSnapResponseError(r.Context(), w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(constant.ErrMsgUnauthorized)))
			return
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateRequestId))
	}

}

func clientErrorHandler(headerKey string, useErrHandler bool) idempotenshine.ResponseHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := r.Header.Get(headerKey)
		if key == "" {
			response.SendOpenApiNonSnapResponseError(r.Context(), w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(constant.ErrMsgUnauthorized)))
			return
		}

		if !useErrHandler {
			err = constant.ErrDuplicateRequestId
		}

		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, err))
	}
}

func serverErrorHandler(headerKey string) idempotenshine.ResponseHandler {
	return func(err error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		key := r.Header.Get(headerKey)
		if key == "" {
			response.SendOpenApiNonSnapResponseError(r.Context(), w, pkgErrs.New(response.HttpErrUnauthorized, fmt.Errorf(constant.ErrMsgUnauthorized)))
			return
		}
		response.SendOpenApiNonSnapResponseError(ctx, w, pkgErrs.New(response.HttpErrRequest, constant.ErrDuplicateRequestId))
	}
}

func isMerchantExcludeKeyValidation(merchant *merchant.MerchantAuthTokenClaims) bool {
	if merchant == nil {
		return false
	}

	featureFlag := ffcontext.NewEvaluationContext(merchant.MerchantId)

	isActive, err := ffclient.BoolVariation("backend-portal-merchant-idempotency-key-validation-exclude", featureFlag, false)
	return isActive && err == nil
}
