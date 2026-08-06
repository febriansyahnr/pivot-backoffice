package qris_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMock "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/qris"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegistration(t *testing.T) {
	service := serviceMock.NewIQrisService(t)
	testConfig := &config.Config{
		Environment: c.EnvironmentStaging, // Use staging environment for tests
	}

	handler := New(validatorExt.New(), service, testConfig)

	router := chi.NewRouter()
	router.Post("/qr/registrations", handler.Registration)

	request := `{"merchantId":"63272b07-d185-45b1-8323-389a045b5ecd", "acquirer":"BNC", "merchant_type":"Merchant", "createdBy": "John Wick"}`

	tests := []struct {
		name           string
		request        string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR:Invalid body format",
			request:        "A",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:           "ERROR:Invalid data",
			request:        `{"merchantId":"", "acquirer":"BNC", "merchant_type":"Merchant", "createdBy": "John Wick"}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantId":"Key: 'RegistrationReq.MerchantId' Error:Field validation for 'MerchantId' failed on the 'required' tag"}}`,
		},
		{
			name:    "ERROR:Some error",
			request: request,
			setupMock: func() {
				service.On(
					"Registration", c.ValueCtxMockType(), mock.AnythingOfType("*qris.RegistrationReq"),
				).Once().Return("", c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:    "SUCCESS",
			request: request,
			setupMock: func() {
				service.On(
					"Registration", c.ValueCtxMockType(), mock.AnythingOfType("*qris.RegistrationReq"),
				).Return("1234", nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":{"id":"1234"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMock != nil {
				test.setupMock()
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/qr/registrations", strings.NewReader(test.request))

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Output:", rec.Body.String())
			}
		})
	}
}
