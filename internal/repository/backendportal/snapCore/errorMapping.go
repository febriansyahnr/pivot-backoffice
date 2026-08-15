package snapCoreRepository

import (
	"net/http"

	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

// mapSnapCoreResponseCodeToErrorType maps snap-core response codes (e.g., "SNP_BAD_GATEWAY")
// to backend-portal error types. This is needed because snap-core normalizes HTTP 502/503/504
// to 408 (to avoid CDN interception), but preserves the real error code in the response body.
func mapSnapCoreResponseCodeToErrorType(code string) string {
	switch code {
	case "SNP_BAD_GATEWAY":
		return httpResponse.HttpErrBadGateway
	case "SNP_SERVICE_UNAVAILABLE", "SNP_CHANNEL_UNAVAILABLE", "SNP_PARTNER_CHANNEL_ERROR":
		return httpResponse.HttpErrServiceUnavailable
	case "SNP_GATEWAY_TIMEOUT":
		return httpResponse.HttpErrRequestTimeout
	case "SNP_FREQUENCY_ABOVE_LIMIT":
		return httpResponse.HttpErrRequestLimitExceeded
	case "SNP_INTERNAL_ERROR", "SNP_DATABASE_ERROR":
		return httpResponse.HttpErrThirdParty
	}
	return ""
}

func mapPartnerHTTPStatusToErrorType(statusCode int) (string, bool) {
	switch statusCode {
	case http.StatusTooManyRequests:
		return httpResponse.HttpErrRequestLimitExceeded, true
	case http.StatusRequestTimeout, http.StatusGatewayTimeout:
		return httpResponse.HttpErrRequestTimeout, true
	case http.StatusBadGateway:
		return httpResponse.HttpErrBadGateway, true
	case http.StatusServiceUnavailable:
		return httpResponse.HttpErrServiceUnavailable, true
	}

	if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
		return httpResponse.HttpErrRequest, true
	}

	if statusCode >= http.StatusInternalServerError {
		return httpResponse.HttpErrThirdParty, true
	}

	return "", false
}
