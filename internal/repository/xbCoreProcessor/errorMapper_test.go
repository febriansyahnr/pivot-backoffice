package xbCoreProcessorRepository

import (
	"testing"

	pkgErrors "github.com/paper-indonesia/pivot-backoffice/pkg/error"
	httpResponse "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
)

func TestMapXbStatusToError(t *testing.T) {
	testCases := []struct {
		name            string
		statusCode      int
		errMsg          string
		expectedErrType string
	}{
		{"400 maps to HttpErrRequest", 400, "validation error", httpResponse.HttpErrRequest},
		{"401 maps to HttpErrUnauthorized", 401, "unauthorized", httpResponse.HttpErrUnauthorized},
		{"403 maps to HttpErrForbidden", 403, "forbidden", httpResponse.HttpErrForbidden},
		{"404 maps to HttpErrNotFound", 404, "not found", httpResponse.HttpErrNotFound},
		{"408 maps to HttpErrThirdParty", 408, "service unavailable", httpResponse.HttpErrThirdParty},
		{"409 maps to HttpErrDupCheck", 409, "duplicate", httpResponse.HttpErrDupCheck},
		{"429 maps to HttpErrTooManyRequest", 429, "rate limited", httpResponse.HttpErrTooManyRequest},
		{"500 maps to HttpErrInternal", 500, "internal error", httpResponse.HttpErrInternal},
		{"502 maps to HttpErrInternal", 502, "bad gateway", httpResponse.HttpErrInternal},
		{"504 maps to HttpErrInternal", 504, "gateway timeout", httpResponse.HttpErrInternal},
		{"422 defaults to HttpErrRequest", 422, "unprocessable", httpResponse.HttpErrRequest},
		{"418 defaults to HttpErrRequest", 418, "teapot", httpResponse.HttpErrRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := mapXbStatusToError(tc.statusCode, tc.errMsg)
			assert.Error(t, err)

			errType, extractedErr := pkgErrors.ExtractError(err)
			assert.Equal(t, tc.expectedErrType, errType)
			assert.Equal(t, tc.errMsg, extractedErr.Error())
		})
	}
}
