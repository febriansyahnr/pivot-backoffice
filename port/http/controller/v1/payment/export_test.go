package payment_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	c "github.com/paper-indonesia/pivot-backoffice/constant"
	paymentModel "github.com/paper-indonesia/pivot-backoffice/internal/model/payment"
	"github.com/paper-indonesia/pivot-backoffice/internal/model/user"
	serviceMocks "github.com/paper-indonesia/pivot-backoffice/mocks/service"
	s "github.com/paper-indonesia/pivot-backoffice/pkg/util/response"
	. "github.com/paper-indonesia/pivot-backoffice/port/http/controller/v1/payment"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestExport(t *testing.T) {
	service := serviceMocks.NewIPaymentService(t)

	handler := New(nil, validator.New(), nil, WithPaymentService(service))

	router := chi.NewRouter()
	router.Post("/payments/export", handler.Export)

	userTokenClaims := &user.UserTokenClaims{
		UUID:       "000c4096-e92e-4f59-a0fe-bab1fd53b1c9",
		MerchantId: "23bd62a9-239f-4d90-a6b0-6bf22fdec793",
	}

	now := time.Now().UTC()
	endDateStr := now.Format(time.DateOnly)
	startDateStr := now.AddDate(0, 0, -10).Format(time.DateOnly)

	tests := []struct {
		name            string
		userTokenClaims *user.UserTokenClaims
		requestBody     string
		setupMock       func()
		wantStatusCode  int
		wantRespBody    string
	}{
		{
			name:           "ERROR:User not found",
			wantStatusCode: http.StatusUnauthorized,
			wantRespBody:   c.WrapErrApiRespForTest(41, s.ErrTypeAPI, "user not found"),
		},
		{
			name:            "ERROR:Invalid request body",
			userTokenClaims: userTokenClaims,
			requestBody:     `A`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    c.WrapErrApiRespForTest(40, s.ErrTypeAPI, "invalid character 'A' looking for beginning of value"),
		},
		{
			name:            "ERROR:Empty end date",
			userTokenClaims: userTokenClaims,
			requestBody:     `{"startDate": "2024-10-01"}`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    `{"code":"40","message":"invalid validation","error":{"type":"API_ERROR","details":[{"field":"EndDate","message":"Key: 'PaymentDownloadHistoryRequest.EndDate' Error:Field validation for 'EndDate' failed on the 'required' tag"}],"traceId":""},"data":null}`,
		},
		{
			name:            "ERROR:Invalid date range input",
			userTokenClaims: userTokenClaims,
			requestBody:     `{"startDate": "2024-01-01", "endDate": "2024-01-31"}`,
			wantStatusCode:  http.StatusBadRequest,
			wantRespBody:    `{"code":"40","message":"The date range exceeds the allowed backdate limit. Maximum allowed is the last 6 months.","error":{"type":"API_ERROR","details":[],"traceId":""},"data":null}`,
		},
		{
			name:            "ERROR:Some error", // NOSONAR
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"Export", c.ValueCtxMockType(), mock.Anything,
				).Once().Return(nil, c.ErrSomeErrorForUnitTest)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantRespBody:   c.WrapErrApiRespForTest(99, s.ErrTypeUnknown, "some error"),
		},
		{
			name:            "SUCCESS",
			userTokenClaims: userTokenClaims,
			requestBody:     fmt.Sprintf(`{"startDate": "%s", "endDate": "%s"}`, startDateStr, endDateStr),
			setupMock: func() {
				service.On(
					"Export", c.ValueCtxMockType(), mock.Anything,
				).Return(&paymentModel.PaymentDownloadHistoryResponse{URL: "https://"}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantRespBody:   `{"code":"00","message":"OK","data":{"url":"https://"}}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/payments/export", strings.NewReader(test.requestBody))

			if test.setupMock != nil {
				test.setupMock()
			}
			if test.userTokenClaims != nil {
				req = req.WithContext(context.WithValue(req.Context(), c.CtxUserInfoKey, test.userTokenClaims))
			}

			router.ServeHTTP(rec, req)
			assert.Equal(t, test.wantStatusCode, rec.Result().StatusCode)
			if !assert.JSONEq(t, test.wantRespBody, rec.Body.String()) {
				t.Log("Result:", rec.Body.String())
			}
		})
	}
}
