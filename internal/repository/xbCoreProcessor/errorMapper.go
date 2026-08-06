package xbCoreProcessorRepository

import (
	"errors"
	"net/http"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
)

func mapXbStatusToError(statusCode int, errMsg string) error {
	var errType string
	switch statusCode {
	case http.StatusBadRequest:
		errType = httpResponse.HttpErrRequest
	case http.StatusUnauthorized:
		errType = httpResponse.HttpErrUnauthorized
	case http.StatusForbidden:
		errType = httpResponse.HttpErrForbidden
	case http.StatusNotFound:
		errType = httpResponse.HttpErrNotFound
	case http.StatusRequestTimeout:
		errType = httpResponse.HttpErrThirdParty
	case http.StatusConflict:
		errType = httpResponse.HttpErrDupCheck
	case http.StatusTooManyRequests:
		errType = httpResponse.HttpErrTooManyRequest
	default:
		if statusCode >= http.StatusInternalServerError {
			errType = httpResponse.HttpErrInternal
		} else {
			errType = httpResponse.HttpErrRequest
		}
	}
	return pkgErrors.New(errType, errors.New(errMsg))
}
