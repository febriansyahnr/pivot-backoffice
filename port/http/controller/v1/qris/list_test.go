package qris_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paper-indonesia/pivot-backoffice/config"
	c "github.com/paper-indonesia/pivot-backoffice/constant"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/qris"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	"github.com/paper-indonesia/pivot-backoffice/pkg/validatorExt"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/qris"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegistrationList(t *testing.T) {
	service := serviceMocks.NewIQrisService(t)
	testConfig := &config.Config{
		Environment: c.EnvironmentStaging, // Use staging environment for tests
	}

	handler := New(validatorExt.New(), service, testConfig)

	router := chi.NewRouter()
	router.Get("/qr/registrations/merchants/{id}", handler.RegistrationList)

	tests := []struct {
		name           string
		merchantId     string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:       "ERROR:Invalid merchant id",
			merchantId: "A",
			setupMock: func() {
				// Empty body function
			},
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "merchant id is not valid"),
		},
		{
			name:       "ERROR:Some error",
			merchantId: uuid.NewString(),
			setupMock: func() {
				service.On(
					"RegistrationList", c.ValueCtxMockType(), c.StringMockType(),
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, c.ErrSomeErrorForUnitTest.Error()),
		},
		{
			name:       "SUCCESS",
			merchantId: uuid.NewString(),
			setupMock: func() {
				service.On(
					"RegistrationList", c.ValueCtxMockType(), c.StringMockType(),
				).Return([]qris.RegistrationListResp{
					{
						Id:                       "ID",
						ExternalId:               "EX",
						Acquirer:                 "AC",
						MerchantType:             "MT",
						AcquirerMerchantParentId: "",
						MerchantName:             "MN",
						Status:                   "FAILED",
					},
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","data":[{"id":"ID","externalId":"EX","acquirer":"AC","merchantType":"MT","acquirerMerchantParentId":"","merchantName":"MN","status":"FAILED","acquirerMerchantId":null,"callbackDetail":null,"callbackDatetime":null,"createdAt":"0001-01-01T00:00:00Z","updatedAt":"0001-01-01T00:00:00Z"}]}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.setupMock()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/qr/registrations/merchants/"+test.merchantId, nil)

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Actual Response:", rec.Body.String())
			}
		})
	}
}
