package withdrawalCrmController_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/crmController/withdrawal"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetryTransaction(t *testing.T) {
	service := serviceMocks.NewIWithdrawalService(t)

	handler := New(validator.New(), service)

	router := chi.NewRouter()
	router.Post("/withdrawals/{id}/retry-transaction", handler.RetryTransaction)

	merchantId := uuid.NewString()
	withdrawalId := uuid.NewString()
	requestBody := fmt.Sprintf(`{"merchantId": "%s"}`, merchantId)

	tests := []struct {
		name           string
		id             string
		requestBody    string
		queryParams    string
		setupMock      func()
		wantStatusCode int
		wantRespBody   string
	}{
		{
			name:           "ERROR: Invalid UUID",
			id:             "not-a-uuid",
			requestBody:    requestBody,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "id is required"),
		},
		{
			name:           "ERROR: Invalid request body",
			id:             withdrawalId,
			requestBody:    "invalid-json",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, "invalid character 'i' looking for beginning of value"),
		},
		{
			name:           "ERROR: Missing merchantId",
			id:             withdrawalId,
			requestBody:    `{}`,
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   `{"code":"40","errors":{"MerchantID":"Key: 'RetryTransactionRequest.MerchantID' Error:Field validation for 'MerchantID' failed on the 'required' tag"}}`,
		},
		{
			name:           "ERROR: Invalid forceFailed query param",
			id:             withdrawalId,
			requestBody:    requestBody,
			queryParams:    "?forceFailed=notbool",
			wantStatusCode: http.StatusBadRequest,
			wantRespBody:   c.WrapErrRespForTest(40, `invalid forceFailed parameter: must be true or false`),
		},
		{
			name: "ERROR: Service returns error",
			id:   withdrawalId,
			setupMock: func() {
				service.On(
					"RetryTransaction", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.RetryTransactionRequest"),
				).Once().Return(c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrRespForTest(99, "some error"),
		},
		{
			name:        "SUCCESS: Basic retry",
			id:          withdrawalId,
			requestBody: requestBody,
			setupMock: func() {
				service.On(
					"RetryTransaction", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.RetryTransactionRequest"),
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s"}}`, withdrawalId),
		},
		{
			name:        "SUCCESS: With forceFailed query param",
			id:          withdrawalId,
			requestBody: requestBody,
			queryParams: "?forceFailed=true",
			setupMock: func() {
				service.On(
					"RetryTransaction", c.ValueCtxMockType(), mock.AnythingOfType("*withdrawal.RetryTransactionRequest"),
				).Once().Return(nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   fmt.Sprintf(`{"code":"00","data":{"id":"%s"}}`, withdrawalId),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.requestBody == "" {
				test.requestBody = requestBody
			}

			url := "/withdrawals/" + test.id + "/retry-transaction" + test.queryParams
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
