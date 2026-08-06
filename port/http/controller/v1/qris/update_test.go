package qris_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/qris"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReuploadDocument(t *testing.T) {
	service := serviceMocks.NewIQrisService(t)
	testConfig := &config.Config{
		Environment: c.EnvironmentStaging, // Use staging environment for tests
	}

	handler := New(validatorExt.New(), service, testConfig)

	router := chi.NewRouter()
	router.Put("/qr/registrations/documents", handler.ReuploadDocument)

	validReqBody := `{"registrationId": "123456","documentType": "CertificateEstablishment"}`

	tests := []struct {
		name           string
		reqBody        string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:    "ERROR:Invalid request body",
			reqBody: `A`,
			setupMock: func() {
				// Empty body function
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:    "ERROR:Invalid registration id",
			reqBody: `{"registrationId": "A","documentType": "CertificateEstablishment"}`,
			setupMock: func() {
				// Empty body function
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"RegistrationId":"Key: 'ReuploadDocumentReq.RegistrationId' Error:Field validation for 'RegistrationId' failed on the 'numeric' tag"}}`,
		},
		{
			name:    "ERROR:Some error",
			reqBody: validReqBody,
			setupMock: func() {
				service.On(
					"ReuploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*qris.ReuploadDocumentReq"),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, c.ErrSomeErrorForUnitTest.Error()),
		},
		{
			name:    "SUCCESS",
			reqBody: validReqBody,
			setupMock: func() {
				service.On(
					"ReuploadDocument", c.ValueCtxMockType(), mock.AnythingOfType("*qris.ReuploadDocumentReq"),
				).Return(&qris.ReuploadDocumentResp{Uploaded: true}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"uploaded":true}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/qr/registrations/documents", strings.NewReader(test.reqBody))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}
