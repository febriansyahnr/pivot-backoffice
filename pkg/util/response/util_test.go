package response_test

import (
	"net/http"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	"github.com/stretchr/testify/assert"
)

func TestHttpStatusErrorCode(t *testing.T) {
	tests := []struct {
		name           string
		errType        string
		wantCode       string
		wantStatusCode int
	}{
		{
			name:           "database",
			errType:        response.HttpErrDatabase,
			wantCode:       response.HttpStatusErrorDatabase,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "third party",
			errType:        response.HttpErrThirdParty,
			wantCode:       response.HttpStatusErrorThirdParty,
			wantStatusCode: http.StatusInternalServerError,
		},
		{
			name:           "not found",
			errType:        response.HttpErrNotFound,
			wantCode:       response.HttpStatusErrorNotFound,
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "unauthorized",
			errType:        response.HttpErrUnauthorized,
			wantCode:       response.HttpStatusErrorUnauthorized,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "forbidden",
			errType:        response.HttpErrForbidden,
			wantCode:       response.HttpStatusErrorForbidden,
			wantStatusCode: http.StatusForbidden,
		},
		{
			name:           "duplicate check",
			errType:        response.HttpErrDupCheck,
			wantCode:       response.HttpStatusErrorDuplicatedCheck,
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "conflict status code constant",
			errType:        response.HttpStatusErrorConflict,
			wantCode:       response.HttpStatusErrorDuplicatedCheck,
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "request",
			errType:        response.HttpErrRequest,
			wantCode:       response.HttpStatusErrorRequest,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "bad gateway",
			errType:        response.HttpErrBadGateway,
			wantCode:       response.HttpStatusErrorBadGateway,
			wantStatusCode: http.StatusBadGateway,
		},
		{
			name:           "service unavailable",
			errType:        response.HttpErrServiceUnavailable,
			wantCode:       response.HttpStatusServiceUnavailable,
			wantStatusCode: http.StatusServiceUnavailable,
		},
		{
			name:           "validation",
			errType:        response.HttpErrValidation,
			wantCode:       response.HttpStatusErrorValidation,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "unprocessable content",
			errType:        response.HttpErrUnprocessableContent,
			wantCode:       response.HttpStatusErrorUnprocessableContent,
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:           "too many request",
			errType:        response.HttpErrTooManyRequest,
			wantCode:       response.HttpStatusErrorUnprocessableContent,
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:           "daily limit reached",
			errType:        response.HttpErrDailyLimitReached,
			wantCode:       response.HttpStatusErrorDailyLimitReached,
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:           "request limit exceeded",
			errType:        response.HttpErrRequestLimitExceeded,
			wantCode:       response.HttpStatusErrorRequestLimitExceeded,
			wantStatusCode: http.StatusTooManyRequests,
		},
		{
			name:           "resource locked",
			errType:        response.HttpErrResourceLocked,
			wantCode:       response.HttpStatusErrorUnprocessableContent,
			wantStatusCode: http.StatusLocked,
		},
		{
			name:           "request timeout",
			errType:        response.HttpErrRequestTimeout,
			wantCode:       response.HttpStatusErrorrRequestTimeout,
			wantStatusCode: http.StatusGatewayTimeout,
		},
		{
			name:           "conflict",
			errType:        response.HttpErrConflict,
			wantCode:       response.HttpStatusErrorConflict,
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "default internal",
			errType:        "UNKNOWN_ERROR_TYPE",
			wantCode:       response.HttpStatusErrorInternal,
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCode, gotStatusCode := response.HttpStatusErrorCode(tt.errType)
			assert.Equal(t, tt.wantCode, gotCode)
			assert.Equal(t, tt.wantStatusCode, gotStatusCode)
		})
	}
}

func TestGetErrorType(t *testing.T) {
	tests := []struct {
		name        string
		errCode     string
		wantErrType string
	}{
		{name: "not found", errCode: response.HttpErrNotFound, wantErrType: response.ErrTypeAPI},
		{name: "unauthorized", errCode: response.HttpErrUnauthorized, wantErrType: response.ErrTypeAPI},
		{name: "forbidden", errCode: response.HttpErrForbidden, wantErrType: response.ErrTypeAPI},
		{name: "duplicate check", errCode: response.HttpErrDupCheck, wantErrType: response.ErrTypeAPI},
		{name: "internal", errCode: response.HttpErrInternal, wantErrType: response.ErrTypeAPI},
		{name: "request", errCode: response.HttpErrRequest, wantErrType: response.ErrTypeAPI},
		{name: "daily limit reached", errCode: response.HttpErrDailyLimitReached, wantErrType: response.ErrTypeAPI},
		{name: "conflict status code", errCode: response.HttpStatusErrorConflict, wantErrType: response.ErrTypeAPI},
		{name: "validation", errCode: response.HttpErrValidation, wantErrType: response.ErrTypeAPIValidation},
		{name: "bad gateway", errCode: response.HttpErrBadGateway, wantErrType: response.ErrTypeGateway},
		{name: "service unavailable", errCode: response.HttpErrServiceUnavailable, wantErrType: response.ErrTypeGateway},
		{name: "request timeout", errCode: response.HttpErrRequestTimeout, wantErrType: response.ErrTypeGateway},
		{name: "third party", errCode: response.HttpErrThirdParty, wantErrType: response.ErrTypePartner},
		{name: "unknown", errCode: "UNKNOWN_ERROR_CODE", wantErrType: response.ErrTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErrType := response.GetErrorType(tt.errCode)
			assert.Equal(t, tt.wantErrType, gotErrType)
		})
	}
}

func TestGetErrorSourceByHttpErrType(t *testing.T) {
	tests := []struct {
		name       string
		errType    string
		wantSource string
	}{
		{name: "request", errType: response.HttpErrRequest, wantSource: response.ErrorSourceUpstream},
		{name: "unprocessable content", errType: response.HttpErrUnprocessableContent, wantSource: response.ErrorSourceUpstream},
		{name: "unauthorized", errType: response.HttpErrUnauthorized, wantSource: response.ErrorSourceUpstream},
		{name: "bad gateway", errType: response.HttpErrBadGateway, wantSource: response.ErrorSourceDownstream},
		{name: "service unavailable", errType: response.HttpErrServiceUnavailable, wantSource: response.ErrorSourceDownstream},
		{name: "request timeout", errType: response.HttpErrRequestTimeout, wantSource: response.ErrorSourceDownstream},
		{name: "too many request", errType: response.HttpErrTooManyRequest, wantSource: response.ErrorSourceUpstream},
		{name: "request limit exceeded", errType: response.HttpErrRequestLimitExceeded, wantSource: response.ErrorSourceDownstream},
		{name: "third party", errType: response.HttpErrThirdParty, wantSource: response.ErrorSourceDownstream},
		{name: "internal", errType: response.HttpErrInternal, wantSource: response.ErrorSourceSystem},
		{name: "unknown", errType: "UNKNOWN", wantSource: response.ErrorSourceSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSource := response.GetErrorSourceByHttpErrType(tt.errType)
			assert.Equal(t, tt.wantSource, gotSource)
		})
	}
}

func TestGetErrorSource(t *testing.T) {
	tests := []struct {
		name       string
		errType    string
		wantSource string
	}{
		{name: "api error", errType: response.ErrTypeAPI, wantSource: "UPSTREAM"},
		{name: "api validation error", errType: response.ErrTypeAPIValidation, wantSource: "UPSTREAM"},
		{name: "partner error", errType: response.ErrTypePartner, wantSource: "DOWNSTREAM"},
		{name: "gateway error", errType: response.ErrTypeGateway, wantSource: "DOWNSTREAM"},
		{name: "unknown", errType: response.ErrTypeUnknown, wantSource: "SYSTEM"},
		{name: "default", errType: "SOMETHING_ELSE", wantSource: "SYSTEM"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSource := response.GetErrorSource(tt.errType)
			assert.Equal(t, tt.wantSource, gotSource)
		})
	}
}
