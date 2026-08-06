package middleware

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/paper-indonesia/pivot-backoffice/constant"
	pkgErrs "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"

	"go.opentelemetry.io/otel"
)

var otelTracer = otel.Tracer("Middleware")

// ResponseWriter is a wrapper around http.ResponseWriter that captures the response body
type ResponseWriter struct {
	http.ResponseWriter
	body   *bytes.Buffer
	Status int
}

func (w ResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteHeader overrides the default WriteHeader method to capture the response status
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.Status = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *ResponseWriter) BodyString() string {
	return rw.body.String()
}

func (rw *ResponseWriter) BodyBytes() []byte {
	return rw.body.Bytes()
}

func GetTimeLocationFromHeader(name string) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loc, err := time.LoadLocation(r.Header.Get(name))
			if err != nil {
				response.SendApiResponseError(r.Context(), w, pkgErrs.New(response.HttpErrRequest, err))
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), constant.CtxTimeLocation, loc))

			next.ServeHTTP(w, r)
		})
	}
}
