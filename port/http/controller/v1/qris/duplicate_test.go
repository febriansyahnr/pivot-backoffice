package qris_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/qris"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDuplicateRegistration(t *testing.T) {
	// Create valid request with valid UUIDs
	validRequest := qris.DuplicateRegistrationReq{
		SourceMerchantId: "63272b07-d185-45b1-8323-389a045b5ecd",
		TargetMerchantId: "73272b07-d185-45b1-8323-389a045b5ecd",
	}
	validRequestBody, _ := json.Marshal(validRequest)

	tests := []struct {
		name           string
		environment    string
		requestBody    []byte
		setupMock      func(service *serviceMock.IQrisService)
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:        "ERROR:Production environment not allowed",
			environment: c.EnvironmentProduction,
			requestBody: validRequestBody,
			setupMock: func(service *serviceMock.IQrisService) {
				// No service calls should happen
			},
			wantStatusCode: http.StatusForbidden,
			wantRespBody:   c.WrapErrRespForTest(43, "forbidden access"),
		},
		{
			name:        "ERROR:Invalid JSON request",
			environment: c.EnvironmentStaging,
			requestBody: []byte("invalid-json"),
			setupMock: func(service *serviceMock.IQrisService) {
				// No service calls should happen
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":"invalid character 'i' looking for beginning of value"}`,
		},
		{
			name:        "ERROR:Invalid request data",
			environment: c.EnvironmentStaging,
			requestBody: []byte(`{"source_merchant_id": ""}`),
			setupMock: func(service *serviceMock.IQrisService) {
				// No service calls should happen
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":`,
		},
		{
			name:        "ERROR:Service returns error",
			environment: c.EnvironmentStaging,
			requestBody: validRequestBody,
			setupMock: func(service *serviceMock.IQrisService) {
				service.On(
					"DuplicateRegistration", c.ValueCtxMockType(), mock.AnythingOfType("*qris.DuplicateRegistrationReq"),
				).Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS",
			environment: c.EnvironmentStaging,
			requestBody: validRequestBody,
			setupMock: func(service *serviceMock.IQrisService) {
				service.On(
					"DuplicateRegistration", c.ValueCtxMockType(), mock.AnythingOfType("*qris.DuplicateRegistrationReq"),
				).Return("duplicated-registration-id", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"duplicated-registration-id"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			service := serviceMock.NewIQrisService(t)
			tt.setupMock(service)

			// Create config with appropriate environment
			testConfig := &config.Config{
				Environment: tt.environment,
			}

			// Create handler and setup router
			handler := New(validatorExt.New(), service, testConfig)
			router := chi.NewRouter()
			router.Post("/duplicate", handler.DuplicateRegistration)

			// Create request and response recorder
			req := httptest.NewRequest(http.MethodPost, "/duplicate", bytes.NewReader(tt.requestBody))
			rec := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(rec, req)

			// Assert response
			assert.Equal(t, tt.wantStatusCode, rec.Code)
			assert.Contains(t, rec.Body.String(), tt.wantRespBody)
		})
	}
}
