package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/paper-indonesia/pivot-backoffice/config"
	"github.com/paper-indonesia/pivot-backoffice/constant"
	errors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func InternalServiceMiddleware(secret *config.Secret) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := func() error {
				_, segment := otelTracer.Start(r.Context(), "http/middleware/InternalServiceMiddleware")
				defer segment.End()

				// check internal api key
				internalApiKey := r.Header.Get("INTERNAL-API-KEY")
				if internalApiKey == "" {
					return errors.New(response.HttpErrUnauthorized, fmt.Errorf(constant.ErrMsgUnauthorized))
				}

				salt := secret.InternalApiKeySecret.Salt
				hashResult := secret.InternalApiKeySecret.HashResult

				// Calculate the hash using SHA256 and the provided salt
				hash := sha256.New()
				hash.Write([]byte(salt + internalApiKey))
				calculatedHash := hex.EncodeToString(hash.Sum(nil))

				if calculatedHash != hashResult {
					return errors.New(response.HttpErrUnauthorized, fmt.Errorf(constant.ErrMsgUnauthorized))
				}

				return nil
			}()

			if err != nil {
				response.SendOpenApiNonSnapResponseError(r.Context(), w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
