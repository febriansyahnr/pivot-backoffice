package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/redisExt"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"github.com/ezraisw/idemgotent"
	"github.com/ezraisw/wracha/adapter/goredis"
	"github.com/ezraisw/wracha/logger/std"
	"github.com/google/uuid"
	chiExtMiddleware "github.com/paper-indonesia/pdk/v2/chiExt/middleware"
	"github.com/paper-indonesia/pdk/v2/idempotenshine"
	"github.com/paper-indonesia/pdk/v2/logger"
)

func IdempotencyMiddleware(redis redisExt.IRedisExt, pathName string) func(next http.Handler) http.Handler {
	conflictResponder := NewConflictResponder()
	conflictResponder.BEPortal()

	// Accessing the request object
	return idemgotent.Middleware(
		pathName,
		idemgotent.WithAdapter(goredis.NewAdapter(redis.Client())),
		idemgotent.WithResponder(conflictResponder),
		idemgotent.WithLogger(std.NewLogger()),
		idemgotent.WithTTL(24*time.Hour),
		idemgotent.WithKeySource(idemgotent.HeaderKeySource("X-Idempotent-Key")),
	)
}

type Responder interface {
	CacheStatusCode() bool
	CacheHeader() bool
	CacheBody() bool
	Respond(http.ResponseWriter, *http.Request, idemgotent.CacheResult)
	BEPortal()
}

type conflictResponder struct {
	BEPortalResponder bool
}

func NewConflictResponder() Responder {
	return &conflictResponder{}
}

func (conflictResponder) CacheStatusCode() bool {
	return false
}

func (conflictResponder) CacheHeader() bool {
	return false
}

func (conflictResponder) CacheBody() bool {
	return false
}

func (rp *conflictResponder) BEPortal() {
	rp.BEPortalResponder = true
}

func (rp conflictResponder) Respond(w http.ResponseWriter, r *http.Request, cr idemgotent.CacheResult) {
	if cr.FromCache {
		if !rp.BEPortalResponder {
			err := pkgErrors.New(response.HttpErrDupCheck, errors.New("duplicate request"))
			response.SendApiResponseError(r.Context(), w, err)
			return
		}

		err := pkgErrors.New(response.HttpErrDupCheck, constant.ErrIdempotencyViolation)
		response.SendApiResponseError(r.Context(), w, err)
		return
	}

	cr.CopyHeaderTo(w, nil)
	w.WriteHeader(cr.Response.GetStatusCode())
	w.Write(cr.Response.GetBody())
}

func IdempotencyApiMiddleware(rdb redisExt.IRedisExt, serviceName, pathName, keySource string) MiddlewareFunc {
	headerKey := idempotenshine.HeaderKeySource(strings.ToUpper(keySource))
	return idempotenshine.Middleware(
		serviceName, pathName,
		idempotenshine.WithTTL(24*time.Hour),
		idempotenshine.WithKeySource(headerKey),
		idempotenshine.WithRedisClient(rdb.Client()),
		idempotenshine.WithLogicOptionUsage(idempotenshine.LogicOption{
			ReturnErrorWhenKeyExists: true,
		}),
		idempotenshine.WithConflictResponder(withApiConflictResponder(keySource)),
		idempotenshine.WithClientResponder(withApiConflictResponder(keySource)),
	)
}

func withApiConflictResponder(headerKey string) idempotenshine.ResponseHandler {
	return func(_ error, w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if r.Header.Get(headerKey) == "" {
			response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdempotencyKeyRequired))
			return
		}
		response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrConflict, constant.ErrIdempotencyKeyExists))
	}
}

func IdempotencyApiMiddlewareWithInvalidateOnError(rdb redisExt.IRedisExt, log logger.ILogger, serviceName, featureName, headerName string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			id := r.Header.Get(headerName)
			if err := uuid.Validate(id); err != nil {
				response.SendApiResponseError(ctx, w, pkgErrors.New(response.HttpErrRequest, constant.ErrIdempotencyViolation))
				return
			}
			if wr, ok := w.(*chiExtMiddleware.ResponseWriter); ok {
				defer func() {
					isAnExcluded := wr.Status == http.StatusOK ||
						(wr.Status == http.StatusConflict && strings.Contains(wr.BodyString(), `Idempotent-Key already exist`))
					if isAnExcluded {
						return
					}
					idempotencyKey := fmt.Sprintf("%s:pdk-idempotency:%s:%s:%s", serviceName, r.Method, featureName, id)
					if err := rdb.Del(ctx, idempotencyKey).Err(); err != nil {
						log.Error(ctx, fmt.Sprintf("Failed to delete idempotency key on %s method %s path %s", featureName, r.Method, r.URL.Path), logger.Error(err))
					}
				}()
			}
			IdempotencyApiMiddleware(rdb, serviceName, featureName, headerName)(next).ServeHTTP(w, r)
		})
	}
}
